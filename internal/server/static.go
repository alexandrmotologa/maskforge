package server

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:dist
var embeddedWebFS embed.FS

// AttachStaticWeb serves the frontend static assets from either embedded FS or local web/dist.
func (s *Server) AttachStaticWeb() {
	// 1. Try local web/dist directory first (great for development & quick reload)
	localDist := filepath.Join(".", "web", "dist")
	if info, err := os.Stat(localDist); err == nil && info.IsDir() {
		fileServer := http.FileServer(http.Dir(localDist))
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				return
			}
			// Check if file exists, else serve index.html for SPA routing
			targetPath := filepath.Join(localDist, filepath.Clean(r.URL.Path))
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(localDist, "index.html"))
				return
			}
			fileServer.ServeHTTP(w, r)
		})
		return
	}

	// 2. Try embedded files
	distFS, err := fs.Sub(embeddedWebFS, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(distFS))
		s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				return
			}
			f, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
			if err != nil {
				// Fallback to index.html for SPA
				indexData, err := fs.ReadFile(distFS, "index.html")
				if err == nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					w.Write(indexData)
					return
				}
			} else {
				_ = f.Close()
			}
			fileServer.ServeHTTP(w, r)
		})
		return
	}

	// 3. Fallback if web studio not built yet
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html><html><body style="font-family:sans-serif;background:#090d16;color:#e2e8f0;padding:2rem;">
<h2>MaskForge Studio</h2>
<p>Frontend assets are not yet compiled. Run <code>npm run build</code> inside the <code>web/</code> directory.</p>
</body></html>`))
	})
}
