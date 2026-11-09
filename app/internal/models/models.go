package models

import "time"

// Pair is an OpenVPN pair: entry server S1 + exit server S2 linked by GRE.
// After bootstrap, S2 is managed only via GRE through S1 (never via public IP).
type Pair struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Session      bool              `json:"session"`
	GRE          bool              `json:"gre"`
	DNS          string            `json:"dns"`
	DNSCustom    []string          `json:"dnsCustom"`
	Route        string            `json:"route"` // split | via1 | all2
	Lists        map[string][]string `json:"lists"`
	Server1      ServerNode        `json:"a"`
	Server2      ServerNode        `json:"b"`
	GRELocal     string            `json:"greLocal"`  // S1 GRE addr
	GRERemote    string            `json:"greRemote"` // S2 GRE addr
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

type ServerNode struct {
	PublicIP string  `json:"ip"`
	SSHPort  string  `json:"port"`
	SSHLogin string  `json:"login"`
	Online   bool    `json:"online"`
	CPU      *float64 `json:"cpu"`
	RAM      *float64 `json:"ram"`
	HDD      *float64 `json:"hdd"`
	Uptime   string  `json:"uptime"`
	// ManagementIP is how we reach this node after bootstrap.
	// For S2 this must be the GRE address, never the public IP.
	ManagementIP string `json:"managementIp,omitempty"`
}

type User struct {
	ID       int64         `json:"id"`
	PairID   string        `json:"pair"`
	Name     string        `json:"name"`
	Comment  string        `json:"comment"`
	Auth     string        `json:"auth"` // file | login
	Limit    int           `json:"limit"`
	Password string        `json:"password,omitempty"`
	Connections []Connection `json:"connections"`
}

type Connection struct {
	Client string `json:"client"`
	VPN    string `json:"vpn"`
	Online string `json:"online"`
	Traffic string `json:"traffic"`
}

type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type InstallRequest struct {
	Name      string         `json:"name"`
	Servers   []InstallServer `json:"servers"` // [0]=S1 [1]=S2
	UserName  string         `json:"userName"`
	UserComment string       `json:"userComment"`
	Auth      string         `json:"auth"`
	Limit     int            `json:"limit"`
	DNS       string         `json:"dns"`
	DNSCustom []string       `json:"dnsCustom"`
	Route     string         `json:"route"`
	DryRun    *bool          `json:"dryRun,omitempty"`
}

type InstallServer struct {
	IP       string `json:"ip"`
	Port     string `json:"port"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type InstallJob struct {
	ID        string    `json:"id"`
	PairName  string    `json:"pairName"`
	Index     int       `json:"index"`
	Steps     []string  `json:"steps"`
	Log       []string  `json:"log"`
	Done      bool      `json:"done"`
	OK        bool      `json:"ok"`
	DryRun    bool      `json:"dryRun"`
	Error     string    `json:"error,omitempty"`
	PairID    string    `json:"pairId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type ApplySettingsRequest struct {
	DNS       string              `json:"dns"`
	DNSCustom []string            `json:"dnsCustom"`
	Route     string              `json:"route"`
	Lists     map[string][]string `json:"lists"`
}

type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Mode    string `json:"mode"`
	Protocol string `json:"protocol"`
	Notes   []string `json:"notes"`
}
