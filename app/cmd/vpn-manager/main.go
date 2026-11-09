package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

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
	noBrowser := flag.Bool("no-browser", false, "do not open the UI in a browser")
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

	url := "http://" + *addr + "/"
	fmt.Println(api.Banner(version, url, map[bool]string{true: "dry-run", false: "live"}[dry]))
	fmt.Println("OpenVPN only · Server 2 admin path = GRE via Server 1")

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	if !*noBrowser {
		go func() {
			// Wait until the listener accepts connections.
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				c, err := net.DialTimeout("tcp", *addr, 200*time.Millisecond)
				if err == nil {
					_ = c.Close()
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			openBrowser(url)
		}()
	}

	log.Fatal(http.Serve(ln, withCORS(root)))
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("open browser: %v", err)
	}
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
