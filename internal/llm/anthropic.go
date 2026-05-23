package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

type Client struct {
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

func NewClient() *Client {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	maxTokens := 400
	if v := os.Getenv("ANTHROPIC_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxTokens = n
		}
	}
	return &Client{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

type messagesRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system"`
	Messages  []messageItem `json:"messages"`
}

type messageItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

const systemPrompt = `Eres el Agente 360 de Banco Serfinanza. Ayudas a ASESORES del banco y a CLIENTES a obtener información de productos y procesos de forma precisa.

REGLAS:
1. Responde ÚNICAMENTE con la información del CONTEXTO de abajo. Si no está, di exactamente: "No tengo esa información en mis fuentes oficiales" y sugiere el canal adecuado o hablar con un asesor. NUNCA inventes tasas, costos, plazos ni fechas.
2. Cita SIEMPRE la fuente al final de cada respuesta de conocimiento.
   - channel "asesor": formato [Fuente: <doc>, <seccion>].
   - channel "cliente"/"whatsapp": cita suave, ej. "(según la información oficial de Serfinanza)".
3. NUNCA pidas OTP, claves, contraseñas ni datos sensibles. Si te los piden/ofrecen, recházalo: "Banco Serfinanza nunca solicita claves ni OTP".
4. Tono claro, en español. "cliente"/"whatsapp": cálido y breve. "asesor": directo y preciso.
5. "superCDT" = CDT Serfinanza (mismo producto).
6. Responde de forma concisa: máximo 150 palabras. Usa listas cortas solo si ayudan.

CONTEXTO:
%s`

func (c *Client) Generate(ctx context.Context, question, channel string, chunks []domain.Chunk) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY no configurada")
	}

	contextText := buildContext(chunks)

	system := fmt.Sprintf(systemPrompt, contextText)

	messages := []messageItem{
		{Role: "user", Content: question},
	}

	body, _ := json.Marshal(messagesRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    system,
		Messages:  messages,
	})

	url := "https://api.anthropic.com/v1/messages"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("anthropic api %d: %s", resp.StatusCode, string(raw))
	}

	var result messagesResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("respuesta vacía del LLM")
	}

	answer := result.Content[0].Text
	answer = addCitation(answer, chunks, channel)

	return answer, nil
}

func buildContext(chunks []domain.Chunk) string {
	if len(chunks) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, c := range chunks {
		sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", c.Doc, c.Seccion, c.Contenido))
	}
	return sb.String()
}

func addCitation(answer string, chunks []domain.Chunk, channel string) string {
	if len(chunks) == 0 {
		return answer
	}
	best := chunks[0]
	var citation string
	switch channel {
	case "asesor":
		citation = fmt.Sprintf(" [Fuente: %s, %s]", best.Doc, best.Seccion)
	case "cliente", "whatsapp":
		citation = fmt.Sprintf(" (según %s)", best.Doc)
	default:
		citation = fmt.Sprintf(" [Fuente: %s]", best.Doc)
	}
	return answer + citation
}