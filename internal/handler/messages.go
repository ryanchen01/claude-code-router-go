package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ryanchen01/claude-code-router-go/internal/api"
)

const (
	defaultAnthropicBaseURL = "https://api.anthropic.com"
	upstreamMessagesPath    = "/v1/messages"

	headerAnthropicVersion = "Anthropic-Version"
	headerAnthropicBeta    = "Anthropic-Beta"
	headerXAPIKey          = "X-Api-Key"
)

var passthroughResponseHeaders = []string{
	headerAnthropicVersion,
	headerAnthropicBeta,
	"Request-Id",
	"X-Request-Id",
	"Retry-After",
}

type MessagesHandler struct {
	client         *http.Client
	baseURL        string
	defaultAPIKey  string
	defaultVersion string
	defaultBeta    []string
	userAgent      string
}

func (h *MessagesHandler) PostV1messages(w http.ResponseWriter, r *http.Request) *api.Response {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errorResponse(http.StatusBadRequest, "failed to read request body")
	}
	defer r.Body.Close()

	streaming := hasStreamingEnabled(body)

	apiKey := headerOrDefault(r.Header, headerXAPIKey, h.defaultAPIKey)
	if apiKey == "" {
		return errorResponse(http.StatusBadRequest, "missing Anthropic API key; set X-Api-Key header or ANTHROPIC_API_KEY")
	}

	version := headerOrDefault(r.Header, headerAnthropicVersion, h.defaultVersion)
	if version == "" {
		return errorResponse(http.StatusBadRequest, "missing Anthropic API version; set Anthropic-Version header or ANTHROPIC_VERSION")
	}

	upstreamURL := h.baseURL + upstreamMessagesPath
	upstreamReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return errorResponse(http.StatusInternalServerError, "failed to construct upstream request")
	}

	upstreamReq.Header.Set(headerXAPIKey, apiKey)
	upstreamReq.Header.Set(headerAnthropicVersion, version)

	betaHeaders := normalizeCSV(r.Header.Values(headerAnthropicBeta))
	if len(betaHeaders) == 0 {
		betaHeaders = h.defaultBeta
	}
	for _, value := range betaHeaders {
		upstreamReq.Header.Add(headerAnthropicBeta, value)
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	upstreamReq.Header.Set("Content-Type", contentType)

	if streaming {
		if accept := r.Header.Get("Accept"); accept != "" {
			upstreamReq.Header.Set("Accept", accept)
		} else {
			upstreamReq.Header.Set("Accept", "text/event-stream")
		}
	}

	if h.userAgent != "" {
		upstreamReq.Header.Set("User-Agent", h.userAgent)
	}

	resp, err := h.client.Do(upstreamReq)
	if err != nil {
		return errorResponse(http.StatusBadGateway, "failed to reach Anthropic Messages API")
	}
	defer resp.Body.Close()

	copyPassthroughHeaders(w.Header(), resp.Header)

	if streaming {
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(resp.StatusCode)
		copyStream(w, resp.Body)
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorResponse(http.StatusBadGateway, "failed to read response from Anthropic Messages API")
	}

	apiResp := api.NewResponse(resp.StatusCode, json.RawMessage(respBody))

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		apiResp.ContentType(ct)
	} else {
		apiResp.ContentType("application/json")
	}

	return apiResp
}

func NewMessagesHandler() *MessagesHandler {
	baseURL := strings.TrimSuffix(os.Getenv("ANTHROPIC_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = defaultAnthropicBaseURL
	}

	return &MessagesHandler{
		client:         &http.Client{},
		baseURL:        baseURL,
		defaultAPIKey:  strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")),
		defaultVersion: strings.TrimSpace(os.Getenv("ANTHROPIC_VERSION")),
		defaultBeta:    normalizeCSV([]string{os.Getenv("ANTHROPIC_BETA")}),
		userAgent:      "claude-code-router-go/0.1",
	}
}

func errorResponse(code int, message string) *api.Response {
	return api.NewResponse(code, map[string]string{"error": message}).ContentType("application/json")
}

func hasStreamingEnabled(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	var payload struct {
		Stream *bool `json:"stream"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	return payload.Stream != nil && *payload.Stream
}

func headerOrDefault(h http.Header, key, fallback string) string {
	if value := strings.TrimSpace(h.Get(key)); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func normalizeCSV(values []string) []string {
	var result []string
	for _, value := range values {
		for _, piece := range strings.Split(value, ",") {
			piece = strings.TrimSpace(piece)
			if piece != "" {
				result = append(result, piece)
			}
		}
	}
	return result
}

func copyPassthroughHeaders(dst, src http.Header) {
	for _, key := range passthroughResponseHeaders {
		dst.Del(key)
		for _, value := range src.Values(key) {
			if value != "" {
				dst.Add(key, value)
			}
		}
	}
}

func copyStream(dst http.ResponseWriter, src io.Reader) {
	flusher, _ := dst.(http.Flusher)
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			return
		}
	}
}
