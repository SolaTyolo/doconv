// Command server exposes HTTP endpoints for Office document conversion.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SolaTyolo/doconv/pkg/doconv"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/convert", handleConvert)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, "POST /convert (multipart: file + format=markdown|text|json)\nGET /health\n")
	})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           logging(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	fmt.Fprintf(os.Stderr, "doconv server listening on %s\n", *addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Fprintf(os.Stderr, "%s %s %s\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	f, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer f.Close()
	format := strings.ToLower(r.FormValue("format"))
	if format == "" {
		format = "markdown"
	}
	payload, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name := "upload.bin"
	if hdr != nil && hdr.Filename != "" {
		name = hdr.Filename
	}
	doc, err := doconv.ParseReader(bytes.NewReader(payload), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var out []byte
	ct := "text/plain; charset=utf-8"
	switch format {
	case "markdown", "md":
		out = []byte(doconv.ToMarkdown(doc, doconv.DefaultMarkdownOptions()))
		ct = "text/markdown; charset=utf-8"
	case "text", "txt":
		out = []byte(doconv.ToPlainText(doc, doconv.DefaultTextOptions()))
	case "json":
		out, err = doconv.ToJSON(doc, doconv.JSONPretty)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ct = "application/json; charset=utf-8"
	default:
		http.Error(w, "format must be markdown, text, or json", http.StatusBadRequest)
		return
	}
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename=%q`, base+"."+extForFormat(format)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func extForFormat(format string) string {
	switch strings.ToLower(format) {
	case "markdown", "md":
		return "md"
	case "text", "txt":
		return "txt"
	case "json":
		return "json"
	default:
		return "txt"
	}
}
