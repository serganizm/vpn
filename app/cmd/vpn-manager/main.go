package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/serganizm/vpn/app/internal/api"
	"github.com/serganizm/vpn/app/internal/install"
	"github.com/serganizm/vpn/app/internal/store"
)

//go:embed all:web
var webFS embed.FS

//go:embed VERSION
var embeddedVersion string

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "listen address")
	mode := flag.String("mode", envOr("VPN_MANAGER_MODE", "dry-run"), "dry-run|live")
	versionFile := flag.String("version-file", "", "path to VERSION file")
	flag.Parse()

	version := strings.TrimSpace(embeddedVersion)
	if version == "" {
		version = "0.1.0"
	}
	if *versionFile != "" {
		version = api.ReadVersion(*versionFile)
	}

	dry := !strings.EqualFold(*mode, "live")
	st := store.New()
	orch := install.New(st, dry)
	apiServer := api.New(st, orch, version, map[bool]string{true: "dry-run", false: "live"}[dry])

	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	root := http.NewServeMux()
	root.Handle("/api/", apiServer.Mux)
	root.Handle("/", http.FileServer(http.FS(sub)))

	fmt.Println(api.Banner(version, "http://"+*addr+"/", map[bool]string{true: "dry-run", false: "live"}[dry]))
	fmt.Println("OpenVPN only · Server 2 admin path = GRE via Server 1")
	log.Fatal(http.ListenAndServe(*addr, withCORS(root)))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
