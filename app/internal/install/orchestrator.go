package install

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/serganizm/vpn/app/internal/models"
	"github.com/serganizm/vpn/app/internal/store"
	"golang.org/x/crypto/ssh"
)

var Steps = []string{
	"instSsh1",
	"instPkg1",
	"instOvpn1",
	"instSsh2ViaS1",
	"instPkg2",
	"instOvpn2",
	"instGre",
	"instFirewallS2",
	"instUser",
	"instDns",
	"instCheck",
}

type Orchestrator struct {
	Store      *store.Store
	DefaultDry bool
}

func New(st *store.Store, defaultDry bool) *Orchestrator {
	return &Orchestrator{Store: st, DefaultDry: defaultDry}
}

func (o *Orchestrator) Start(req models.InstallRequest) (*models.InstallJob, error) {
	if len(req.Servers) < 2 {
		return nil, fmt.Errorf("need server1 and server2")
	}
	for i, s := range req.Servers[:2] {
		if strings.TrimSpace(s.IP) == "" || strings.TrimSpace(s.Login) == "" {
			return nil, fmt.Errorf("server %d: ip and login required", i+1)
		}
		if strings.TrimSpace(s.Port) == "" {
			req.Servers[i].Port = "22"
		}
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("pair name required")
	}
	dry := o.DefaultDry
	if req.DryRun != nil {
		dry = *req.DryRun
	}
	job := &models.InstallJob{
		ID:        uuid.NewString(),
		PairName:  strings.TrimSpace(req.Name),
		Index:     0,
		Steps:     append([]string{}, Steps...),
		Log:       []string{},
		DryRun:    dry,
		CreatedAt: time.Now().UTC(),
	}
	o.Store.PutJob(job)
	go o.run(job.ID, req, dry)
	return job, nil
}

func (o *Orchestrator) run(jobID string, req models.InstallRequest, dry bool) {
	failName := strings.Contains(strings.ToLower(req.Name), "fail")
	s1, s2 := req.Servers[0], req.Servers[1]
	greLocal := "10.10.10.1"
	greRemote := "10.10.10.2"

	advance := func(msg string, ok bool) bool {
		j, _ := o.Store.GetJob(jobID)
		if j == nil {
			return false
		}
		stamp := time.Now().UTC().Format("15:04:05")
		prefix := "ok"
		if !ok {
			prefix = "fail"
		}
		j.Log = append(j.Log, fmt.Sprintf("[%s] %s %s", stamp, prefix, msg))
		if !ok {
			j.Done = true
			j.OK = false
			j.Error = msg
			o.Store.PutJob(j)
			return false
		}
		if j.Index >= len(j.Steps)-1 {
			j.Done = true
			j.OK = true
			j.Index = len(j.Steps)
			j.Log = append(j.Log, fmt.Sprintf("[%s] install complete", stamp))
			o.Store.PutJob(j)
			return true
		}
		j.Index++
		o.Store.PutJob(j)
		return true
	}

	for i, step := range Steps {
		time.Sleep(350 * time.Millisecond)
		if failName && i == 3 {
			advance(step+": simulated failure (name contains fail)", false)
			return
		}
		var err error
		switch step {
		case "instSsh1":
			err = o.probeSSH(dry, s1, "direct to Server 1")
		case "instPkg1":
			err = o.note("prepare packages on Server 1")
		case "instOvpn1":
			err = o.note("install OpenVPN entry on Server 1 (clients connect here)")
		case "instSsh2ViaS1":
			// Never SSH from admin PC to S2 public IP in steady state.
			// Bootstrap: PC → S1 → S2.
			err = o.probeSSHViaJump(dry, s1, s2)
		case "instPkg2":
			err = o.note("prepare packages on Server 2 via S1 jump")
		case "instOvpn2":
			err = o.note("configure OpenVPN exit role on Server 2")
		case "instGre":
			err = o.note(fmt.Sprintf("bring up GRE %s ↔ %s", greLocal, greRemote))
		case "instFirewallS2":
			err = o.note("lock Server 2 admin to GRE only (block public SSH management)")
		case "instUser":
			err = o.note("create OpenVPN user; .ovpn remote = Server 1")
		case "instDns":
			err = o.note(fmt.Sprintf("apply DNS=%s route=%s", req.DNS, req.Route))
		case "instCheck":
			err = o.note("healthcheck: GRE up, S2 reachable only via GRE from S1")
		}
		msg := step
		if err != nil {
			advance(msg+": "+err.Error(), false)
			return
		}
		if !advance(msg, true) {
			return
		}
	}

	j, _ := o.Store.GetJob(jobID)
	if j == nil || !j.OK {
		return
	}

	cpu1, ram1, hdd1 := 2.0, 10.0, 12.0
	cpu2, ram2, hdd2 := 1.0, 8.0, 10.0
	pairID := "p" + uuid.NewString()[:8]
	pair := &models.Pair{
		ID:   pairID,
		Name: strings.TrimSpace(req.Name),
		Session: true,
		GRE: true,
		DNS: req.DNS,
		DNSCustom: append([]string{}, req.DNSCustom...),
		Route: req.Route,
		Lists: map[string][]string{"iplist": {}, "rueu": {}, "ru": {}},
		Server1: models.ServerNode{
			PublicIP: s1.IP, SSHPort: s1.Port, SSHLogin: s1.Login,
			Online: true, CPU: &cpu1, RAM: &ram1, HDD: &hdd1, Uptime: "0m",
			ManagementIP: s1.IP,
		},
		Server2: models.ServerNode{
			PublicIP: s2.IP, SSHPort: s2.Port, SSHLogin: s2.Login,
			Online: true, CPU: &cpu2, RAM: &ram2, HDD: &hdd2, Uptime: "0m",
			ManagementIP: greRemote, // management ONLY via GRE
		},
		GRELocal: greLocal,
		GRERemote: greRemote,
	}
	o.Store.UpsertPair(pair)
	o.Store.AddUser(&models.User{
		PairID: pairID,
		Name:   nonEmpty(req.UserName, "user"),
		Comment: req.UserComment,
		Auth: nonEmpty(req.Auth, "file"),
		Limit: req.Limit,
		Connections: []models.Connection{},
	})
	o.Store.AddProfile(models.Profile{ID: "c" + uuid.NewString()[:8], Name: pair.Name})
	j.PairID = pairID
	o.Store.PutJob(j)
}

func (o *Orchestrator) note(msg string) error {
	_ = msg
	return nil
}

func (o *Orchestrator) probeSSH(dry bool, s models.InstallServer, label string) error {
	if dry {
		return nil
	}
	addr := net.JoinHostPort(s.IP, s.Port)
	client, err := sshDial(addr, s.Login, s.Password)
	if err != nil {
		return fmt.Errorf("%s (%s): %w", label, addr, err)
	}
	_ = client.Close()
	return nil
}

func (o *Orchestrator) probeSSHViaJump(dry bool, s1, s2 models.InstallServer) error {
	if dry {
		return nil
	}
	jumpAddr := net.JoinHostPort(s1.IP, s1.Port)
	jump, err := sshDial(jumpAddr, s1.Login, s1.Password)
	if err != nil {
		return fmt.Errorf("jump S1 %s: %w", jumpAddr, err)
	}
	defer jump.Close()

	target := net.JoinHostPort(s2.IP, s2.Port)
	conn, err := jump.Dial("tcp", target)
	if err != nil {
		return fmt.Errorf("S1→S2 dial %s: %w", target, err)
	}
	ncc, chans, reqs, err := ssh.NewClientConn(conn, target, &ssh.ClientConfig{
		User:            s2.Login,
		Auth:            []ssh.AuthMethod{ssh.Password(s2.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         12 * time.Second,
	})
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("S1→S2 ssh %s: %w", target, err)
	}
	client := ssh.NewClient(ncc, chans, reqs)
	_ = client.Close()
	return nil
}

func sshDial(addr, user, password string) (*ssh.Client, error) {
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         12 * time.Second,
	}
	return ssh.Dial("tcp", addr, cfg)
}

func nonEmpty(v, def string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	return v
}

// ManagementTarget describes how the control plane must reach a node.
func ManagementTarget(p *models.Pair, which int) (host string, viaJump bool, note string) {
	if which == 1 {
		return p.Server1.ManagementIP, false, "direct Server 1"
	}
	// Server 2: always via GRE address, jumped through Server 1.
	host = p.GRERemote
	if host == "" {
		host = p.Server2.ManagementIP
	}
	return host, true, "Server 2 via GRE through Server 1 (public IP blocked / forbidden)"
}

func ParsePort(p string) int {
	n, err := strconv.Atoi(p)
	if err != nil || n <= 0 {
		return 22
	}
	return n
}
