package upstream

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientConfig struct {
	BaseURL     string
	BearerToken string
	HTTPClient  *http.Client
}

type Client struct {
	baseURL     string
	bearerToken string
	httpClient  *http.Client
}

func New(cfg ClientConfig) *Client {
	return &Client{
		baseURL:     strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		bearerToken: strings.TrimSpace(cfg.BearerToken),
		httpClient:  httpClient(cfg.HTTPClient),
	}
}

func httpClient(client *http.Client) *http.Client {
	if client != nil {
		return client
	}
	return &http.Client{Timeout: 5 * time.Second}
}

func (c *Client) Configured() bool {
	return c != nil && c.baseURL != ""
}

func (c *Client) Forward(w http.ResponseWriter, r *http.Request, method string, targetPath string) {
	if !c.Configured() {
		writeError(w, http.StatusBadGateway, "upstream_not_configured")
		return
	}
	target, err := c.targetURL(targetPath, r.URL.RawQuery)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_invalid_url")
		return
	}
	request, err := http.NewRequestWithContext(r.Context(), method, target, r.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_request_failed")
		return
	}
	copyForwardHeaders(request.Header, r.Header)
	if c.bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	if request.Header.Get("Idempotency-Key") == "" && method != http.MethodGet {
		request.Header.Set("Idempotency-Key", "console-"+randomToken())
	}
	if request.Header.Get("X-Correlation-Id") == "" {
		request.Header.Set("X-Correlation-Id", "console-"+randomToken())
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		writeError(w, http.StatusBadGateway, "upstream_unavailable")
		return
	}
	defer response.Body.Close()

	copyResponseHeaders(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

func (c *Client) Check(ctx context.Context, path string) bool {
	if !c.Configured() {
		return false
	}
	target, err := c.targetURL(path, "")
	if err != nil {
		return false
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false
	}
	if c.bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices
}

func (c *Client) targetURL(path string, rawQuery string) (string, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	relative, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	target := base.ResolveReference(relative)
	if target.RawQuery == "" {
		target.RawQuery = rawQuery
	}
	return target.String(), nil
}

func copyForwardHeaders(target http.Header, source http.Header) {
	for _, key := range []string{"Content-Type", "Accept", "Idempotency-Key", "X-Correlation-Id"} {
		if value := source.Get(key); value != "" {
			target.Set(key, value)
		}
	}
}

func copyResponseHeaders(target http.Header, source http.Header) {
	if value := source.Get("Content-Type"); value != "" {
		target.Set("Content-Type", value)
	}
}

func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

func randomToken() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return time.Now().UTC().Format("20060102150405")
	}
	return hex.EncodeToString(bytes[:])
}

func JSONBody(value any) io.Reader {
	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(value)
	return &body
}
