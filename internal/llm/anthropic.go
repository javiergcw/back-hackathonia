package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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
1. Responde ÚNICAMENTE con la información del CONTEXTO de abajo. Si no está, di exactamente: "No tengo esa información en mis fuentes oficiales" y sugiere hablar con un asesor. NUNCA inventes tasas, costos, plazos ni fechas.
2. NUNCA pidas OTP, claves, contraseñas ni datos sensibles. Si te los piden/ofrecen, recházalo: "Banco Serfinanza nunca solicita claves ni OTP".
3. "superCDT" = CDT Serfinanza (mismo producto).

CONTEXTO:
%s`

const whatsappSystemPrompt = `Eres el asesor virtual de Banco Serfinanza en WhatsApp. Hablas directamente con el cliente.

Tu misión: orientarlo con claridad, calidez y confianza, como un buen asesor del banco.

REGLAS:
1. El CONTEXTO de abajo es la documentación oficial vigente de Serfinanza. Preséntala con seguridad: afirma los datos, no los supongas ni los cuestiones.
2. NO digas "según el PDF", "según la información oficial", "según Serfinanza" ni nombres de archivos (.pdf). El cliente no necesita ver referencias internas.
3. NO uses frases dubitativas ("podría", "quizás", "creo que"). Si está en el contexto, dilo con convicción.
4. Si la información NO está en el contexto, dilo con honestidad: "No tengo ese dato a la mano, pero un asesor de Serfinanza puede ayudarte" y sugiere el canal adecuado.
5. NUNCA inventes tasas, costos, plazos ni fechas. NUNCA pidas OTP, claves ni datos sensibles.
6. Tono: cercano, profesional, en segunda persona ("te", "tu"). Como un asesor que quiere ayudar de verdad.
7. "superCDT" = CDT Serfinanza (mismo producto).
8. Un solo párrafo fluido, máximo 2-4 oraciones. Sin saltos de línea, listas ni viñetas. Énfasis con *negrita* (formato WhatsApp) solo si aporta.

CONTEXTO:
%s`

func (c *Client) Generate(ctx context.Context, question, channel string, chunks []domain.Chunk) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY no configurada")
	}

	contextText := buildContext(chunks)

	var system string
	if channel == "whatsapp" {
		system = fmt.Sprintf(whatsappSystemPrompt, contextText)
	} else {
		system = fmt.Sprintf(systemPrompt, contextText)
	}

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
	if channel != "whatsapp" {
		answer = addCitation(answer, chunks, channel)
	}

	return answer, nil
}

var (
	markdownBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	citationTail    = regexp.MustCompile(`(?i)\s*\(según[^)]*\)\s*$`)
	pdfCitationTail = regexp.MustCompile(`(?i)\s*\(según\s+[\w._-]+\.pdf\)\s*$`)
)

func FormatForWhatsApp(text string) string {
	text = strings.ReplaceAll(text, "\\n", " ")
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = markdownBold.ReplaceAllString(text, "*$1*")
	text = strings.ReplaceAll(text, "• ", "")
	text = strings.ReplaceAll(text, "- ", "")
	text = strings.ReplaceAll(text, "(según la información oficial de Serfinanza)", "")
	text = strings.ReplaceAll(text, "(según la información oficial de serfinanza)", "")
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	for {
		cleaned := citationTail.ReplaceAllString(text, "")
		cleaned = pdfCitationTail.ReplaceAllString(cleaned, "")
		if cleaned == text {
			break
		}
		text = cleaned
	}
	return strings.TrimSpace(text)
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