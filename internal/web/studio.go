package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// StudioBackend decouples the web server's Studio Kit endpoints (/api/studio/*)
// from pkg/studio so internal/web depends only on the Go standard library and pkg/model.
type StudioBackend interface {
	// PresetsJSON returns JSON of the form {"presets":[...]}.
	PresetsJSON() ([]byte, error)
	// EvaluateJSON accepts a StudioProject JSON payload and returns a StudioEvaluation JSON payload.
	EvaluateJSON(reqBody []byte) ([]byte, error)
	// BuildPacketZip accepts a StudioProject JSON payload and returns a complete build packet
	// .zip archive along with a suggested download filename (e.g. "kilnscrub-wood_salt_24-packet.zip").
	BuildPacketZip(reqBody []byte) ([]byte, string, error)
}

var studioDisabledPresetsJSON = []byte(`{"enabled":false,"presets":[]}`)

func (s *Server) handleStudioPresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	b := s.getStudioBackend()
	if b == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = w.Write(studioDisabledPresetsJSON)
		}
		return
	}

	data, err := b.PresetsJSON()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to load studio presets: %v", err))
		return
	}
	if !json.Valid(data) {
		writeJSONError(w, http.StatusInternalServerError, "studio presets backend returned invalid JSON")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = w.Write(data)
	}
}

func (s *Server) handleStudioEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	b := s.getStudioBackend()
	if b == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "studio backend is not enabled")
		return
	}

	body, err := readLimitedBody(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		writeJSONError(w, http.StatusBadRequest, "request body must be a valid JSON object")
		return
	}

	out, err := b.EvaluateJSON(trimmed)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !json.Valid(out) {
		writeJSONError(w, http.StatusInternalServerError, "studio evaluation backend returned invalid JSON")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

func (s *Server) handleStudioPacket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	b := s.getStudioBackend()
	if b == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "studio backend is not enabled")
		return
	}

	body, err := readLimitedBody(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		writeJSONError(w, http.StatusBadRequest, "request body must be a valid JSON object")
		return
	}

	zipBytes, filename, err := b.BuildPacketZip(trimmed)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	safeName := sanitizeDownloadFilename(filename, "kilnscrub-build-packet.zip")
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, safeName))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}

func sanitizeDownloadFilename(name, fallback string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(name) {
		if r < 0x20 || r == 0x7f || r == '"' || r == '/' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}
	cleaned := strings.TrimLeft(strings.TrimSpace(b.String()), ".")
	if cleaned == "" {
		return fallback
	}
	return cleaned
}
