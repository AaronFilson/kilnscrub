package web

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AaronFilson/kilnscrub/pkg/model"
)

//go:embed static/index.html
var staticFS embed.FS

const (
	maxRequestBodyBytes = 1 << 20 // 1 MiB
	indexHTMLPath       = "static/index.html"
)

// Server serves the embedded kilnscrub engineering dashboard and JSON/CAD HTTP API.
// It is decoupled from solver and CAD implementations via model.Designer and model.CADExporter.
type Server struct {
	designer  model.Designer
	exporter  model.CADExporter
	indexHTML []byte
	mux       *http.ServeMux
}

// NewServer creates a new Server backed by the provided Designer and CADExporter.
func NewServer(designer model.Designer, exporter model.CADExporter) *Server {
	htmlBytes, err := staticFS.ReadFile(indexHTMLPath)
	if err != nil {
		// Unreachable because static/index.html is embedded at compile time.
		panic(fmt.Sprintf("web: failed to read embedded %s: %v", indexHTMLPath, err))
	}

	s := &Server{
		designer:  designer,
		exporter:  exporter,
		indexHTML: htmlBytes,
		mux:       http.NewServeMux(),
	}

	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/defaults", s.handleDefaults)
	s.mux.HandleFunc("/api/size", s.handleSize)
	s.mux.HandleFunc("/api/export", s.handleExport)

	return s
}

// Handler returns the HTTP handler for the web dashboard and API routes.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// ListenAndServe starts the HTTP server on addr and gracefully shuts down when ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("web: listen %s: %w", addr, err)
	}

	httpSrv := &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.Serve(ln)
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr := httpSrv.Shutdown(shutdownCtx)
		serveErr := <-errCh
		if shutdownErr != nil {
			return shutdownErr
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = w.Write(s.indexHTML)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDefaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, model.DefaultSaltKilnInput())
}

func (s *Server) handleSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.designer == nil {
		writeJSONError(w, http.StatusInternalServerError, "designer service not configured")
		return
	}

	body, err := readLimitedBody(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	input, _, err := decodePartialDesignInput(body, false)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := input.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid design input: %v", err))
		return
	}

	design, err := s.designer.Design(input)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("design solver failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, design)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if s.designer == nil || s.exporter == nil {
		writeJSONError(w, http.StatusInternalServerError, "designer or exporter service not configured")
		return
	}

	body, err := readLimitedBody(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	input, bodyFormat, err := decodePartialDesignInput(body, true)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = strings.ToLower(strings.TrimSpace(bodyFormat))
	}

	var contentType, filename string
	switch format {
	case "stl":
		contentType = "model/stl"
		filename = "kilnscrub.stl"
	case "dxf":
		contentType = "application/dxf"
		filename = "kilnscrub.dxf"
	case "bom_csv":
		contentType = "text/csv; charset=utf-8"
		filename = "kilnscrub-bom.csv"
	case "bom_md":
		contentType = "text/markdown; charset=utf-8"
		filename = "kilnscrub-bom.md"
	case "bom_json":
		contentType = "application/json; charset=utf-8"
		filename = "kilnscrub-bom.json"
	default:
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("unsupported export format %q; expected one of stl, dxf, bom_csv, bom_md, bom_json", format))
		return
	}

	if err := input.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid design input: %v", err))
		return
	}

	design, err := s.designer.Design(input)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("design solver failed: %v", err))
		return
	}

	var payload []byte
	switch format {
	case "stl":
		payload, err = s.exporter.ExportSTL(design)
	case "dxf":
		payload, err = s.exporter.ExportDXF(design)
	case "bom_csv":
		var bom model.BOM
		bom, err = s.exporter.GenerateBOM(design)
		if err == nil {
			payload, err = s.exporter.FormatBOMCSV(bom)
		}
	case "bom_md":
		var bom model.BOM
		bom, err = s.exporter.GenerateBOM(design)
		if err == nil {
			payload, err = s.exporter.FormatBOMMarkdown(bom)
		}
	case "bom_json":
		var bom model.BOM
		bom, err = s.exporter.GenerateBOM(design)
		if err == nil {
			payload, err = json.MarshalIndent(bom, "", "  ")
		}
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("export %s failed: %v", format, err))
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func readLimitedBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	if len(data) > maxRequestBodyBytes {
		return nil, fmt.Errorf("request body exceeds maximum size of %d bytes", maxRequestBodyBytes)
	}
	return data, nil
}

// decodePartialDesignInput initializes model.DefaultSaltKilnInput() BEFORE unmarshaling JSON
// so that partial JSON payloads inherit valid engineering defaults for omitted fields.
func decodePartialDesignInput(body []byte, allowEmpty bool) (model.DesignInput, string, error) {
	input := model.DefaultSaltKilnInput()
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		if allowEmpty {
			return input, "", nil
		}
		return input, "", errors.New("request body must be a JSON object")
	}
	if trimmed[0] != '{' {
		return input, "", errors.New("request body must be a JSON object")
	}
	if err := json.Unmarshal(trimmed, &input); err != nil {
		return input, "", fmt.Errorf("invalid JSON payload: %w", err)
	}

	var meta struct {
		Format string `json:"format"`
	}
	_ = json.Unmarshal(trimmed, &meta)

	return input, meta.Format, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, `{"error":"failed to marshal JSON response"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
