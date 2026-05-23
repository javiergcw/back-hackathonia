package evolution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/javierg/hackathon-bqia/internal/infrastructure/config"
)

type Client struct {
	baseURL    string
	apiKey     string
	instance   string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL:  strings.TrimRight(cfg.EvolutionAPIURL, "/"),
		apiKey:   cfg.EvolutionAPIKey,
		instance: cfg.EvolutionInstance,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type SendTextRequest struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

type WebhookSetRequest struct {
	Webhook WebhookConfig `json:"webhook"`
}

type WebhookConfig struct {
	Enabled  bool              `json:"enabled"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers,omitempty"`
	ByEvents bool              `json:"byEvents"`
	Base64   bool              `json:"base64"`
	Events   []string          `json:"events"`
}

func (c *Client) SendText(ctx context.Context, number, text string) (json.RawMessage, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("EVOLUTION_API_KEY no configurada")
	}

	number = NormalizeNumber(number)
	if number == "" {
		return nil, fmt.Errorf("número inválido")
	}

	body, err := json.Marshal(SendTextRequest{Number: number, Text: text})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/message/sendText/%s", c.baseURL, c.instance)
	return c.do(ctx, http.MethodPost, url, body)
}

func (c *Client) SetWebhook(ctx context.Context, webhookURL string) (json.RawMessage, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("EVOLUTION_API_KEY no configurada")
	}

	payload := WebhookSetRequest{
		Webhook: WebhookConfig{
			Enabled:  true,
			URL:      webhookURL,
			ByEvents: false,
			Base64:   false,
			Events: []string{
				"MESSAGES_UPSERT",
				"CONNECTION_UPDATE",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/webhook/set/%s", c.baseURL, c.instance)
	return c.do(ctx, http.MethodPost, url, body)
}

func (c *Client) do(ctx context.Context, method, url string, body []byte) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("evolution api %d: %s", res.StatusCode, string(raw))
	}

	return raw, nil
}

func NormalizeNumber(number string) string {
	number = strings.TrimSpace(number)
	number = strings.TrimPrefix(number, "+")
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")
	if idx := strings.Index(number, "@"); idx > 0 {
		number = number[:idx]
	}
	return number
}
