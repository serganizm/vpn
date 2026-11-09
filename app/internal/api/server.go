package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/serganizm/vpn/app/internal/install"
	"github.com/serganizm/vpn/app/internal/models"
	"github.com/serganizm/vpn/app/internal/openvpn"
	"github.com/serganizm/vpn/app/internal/store"
)

type Server struct {
	Store   *store.Store
	Install *install.Orchestrator
	Version string
	Mode    string
	Mux     *http.ServeMux
}

func New(st *store.Store, orch *install.Orchestrator, version, mode string) *Server {
	s := &Server{Store: st, Install: orch, Version: version, Mode: mode, Mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.Mux.HandleFunc("/api/v1/info", s.handleInfo)
	s.Mux.HandleFunc("/api/v1/pairs", s.handlePairs)
	s.Mux.HandleFunc("/api/v1/pairs/", s.handlePairSub)
	s.Mux.HandleFunc("/api/v1/install", s.handleInstall)
	s.Mux.HandleFunc("/api/v1/install/", s.handleInstallJob)
	s.Mux.HandleFunc("/api/v1/users", s.handleUsers)
	s.Mux.HandleFunc("/api/v1/users/", s.handleUserSub)
	s.Mux.HandleFunc("/api/v1/profiles", s.handleProfiles)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, models.AppInfo{
		Name:     "VPN Manager",
		Version:  s.Version,
		Mode:     s.Mode,
		Protocol: "OpenVPN",
		Notes: []string{
			"Clients connect to Server 1 only",
			"Server 2 is managed via GRE through Server 1",
			"Public IP of Server 2 is never used for admin after bootstrap",
		},
	})
}

func (s *Server) handlePairs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.Store.ListPairs())
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handlePairSub(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/pairs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	pair, ok := s.Store.GetPair(id)
	if !ok {
		http.Error(w, "pair not found", http.StatusNotFound)
		return
	}

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			writeJSON(w, pair)
			return
		}
		methodNotAllowed(w)
		return
	}

	switch parts[1] {
	case "session":
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		var body struct {
			On bool `json:"on"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.Store.SetPairSession(id, body.On)
		pair, _ = s.Store.GetPair(id)
		writeJSON(w, pair)
	case "settings":
		if r.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}
		// Settings for S2 are applied only through S1 → GRE.
		if !pair.Session {
			http.Error(w, "connect to Server 1 session first", http.StatusConflict)
			return
		}
		if !pair.GRE {
			http.Error(w, "GRE is down; cannot manage Server 2 via tunnel", http.StatusConflict)
			return
		}
		var body models.ApplySettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pair.DNS = body.DNS
		pair.DNSCustom = body.DNSCustom
		pair.Route = body.Route
		if body.Lists != nil {
			pair.Lists = body.Lists
		}
		host, viaJump, note := install.ManagementTarget(pair, 2)
		s.Store.UpsertPair(pair)
		writeJSON(w, map[string]any{
			"pair":             pair,
			"appliedVia":       host,
			"viaJump":          viaJump,
			"managementNote":   note,
			"server2PublicIP":  pair.Server2.PublicIP,
			"forbiddenDirect":  true,
		})
	case "management":
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		h1, _, n1 := install.ManagementTarget(pair, 1)
		h2, j2, n2 := install.ManagementTarget(pair, 2)
		writeJSON(w, map[string]any{
			"server1": map[string]any{"host": h1, "viaJump": false, "note": n1},
			"server2": map[string]any{"host": h2, "viaJump": j2, "note": n2, "publicIpForbidden": true},
			"gre":     map[string]any{"local": pair.GRELocal, "remote": pair.GRERemote, "up": pair.GRE},
		})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req models.InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	job, err := s.Install.Start(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, job)
}

func (s *Server) handleInstallJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/install/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	job, ok := s.Store.GetJob(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, job)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pairID := r.URL.Query().Get("pair")
		writeJSON(w, s.Store.ListUsers(pairID))
	case http.MethodPost:
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if u.PairID == "" {
			http.Error(w, "pair required", http.StatusBadRequest)
			return
		}
		pair, ok := s.Store.GetPair(u.PairID)
		if !ok {
			http.Error(w, "pair not found", http.StatusNotFound)
			return
		}
		if !pair.Session || !pair.GRE {
			http.Error(w, "Server 1 session + GRE required to manage users on Server 2 path", http.StatusConflict)
			return
		}
		writeJSON(w, s.Store.AddUser(&u))
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleUserSub(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "bad user id", http.StatusBadRequest)
		return
	}
	user, ok := s.Store.GetUser(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, user)
		case http.MethodPut:
			var body models.User
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			body.ID = id
			body.PairID = user.PairID
			if !s.Store.UpdateUser(&body) {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, body)
		case http.MethodDelete:
			s.Store.DeleteUser(id)
			w.WriteHeader(http.StatusNoContent)
		default:
			methodNotAllowed(w)
		}
		return
	}

	if parts[1] == "config.ovpn" && r.Method == http.MethodGet {
		pair, ok := s.Store.GetPair(user.PairID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		cfg := openvpn.ClientConfig(pair.Server1.PublicIP, 1194, user.Name)
		if openvpn.ContainsForbiddenRemote(cfg, pair.Server2.PublicIP) {
			http.Error(w, "refusing config that points at Server 2 public IP", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-openvpn-profile")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", user.Name+".ovpn"))
		_, _ = io.WriteString(w, cfg)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.Store.ListProfiles())
	case http.MethodPost:
		var p models.Profile
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if p.ID == "" {
			p.ID = "c" + uuid.NewString()[:8]
		}
		s.Store.AddProfile(p)
		writeJSON(w, p)
	default:
		methodNotAllowed(w)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func ReadVersion(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return "0.1.0"
	}
	return strings.TrimSpace(string(b))
}

func Banner(version, addr, mode string) string {
	return fmt.Sprintf("VPN Manager %s · OpenVPN · mode=%s · %s · %s", version, mode, addr, time.Now().Format(time.RFC3339))
}
