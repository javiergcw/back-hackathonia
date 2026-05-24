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
	"github.com/javierg/hackathon-bqia/internal/session"
)

type Client struct {
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

type ClientRequest struct {
	Question       string
	Chunks         []domain.Chunk
	ProfileContext string
	ProactiveHint  string
	History        []session.Message
}

func NewClient() *Client {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-haiku-4-5"
	}
	maxTokens := 1024
	if v := os.Getenv("ANTHROPIC_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxTokens = n
		}
	}
	timeout := 40 * time.Second
	if v := os.Getenv("ANTHROPIC_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}
	return &Client{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		client:    &http.Client{Timeout: timeout},
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

const clientSystemPrompt = `Eres el asesor virtual de Banco Serfinanza: un chatbot con IA cercano, humano y útil. Hablas con un CLIENTE.

PERSONALIDAD:
- Tono casual y cálido, como un buen asesor por chat (no un robot ni un formulario).
- Saludos, despedidas y agradecimientos: respóndelos con naturalidad y brevedad.
- Si el mensaje es vago ("ayúdame", "tengo una duda"), invita con amabilidad a contar qué necesita (CDT, tarjeta, crédito, App, pagos, etc.).

USO DEL CONTEXTO (documentación Serfinanza):
1. Cuando el CONTEXTO tenga datos, úsalos con seguridad. No digas "según el PDF" ni cites archivos.
2. Si el CONTEXTO está vacío o no cubre el detalle exacto, NO digas "no tengo información" ni cierres la conversación. Orienta con canales (App, Web, WhatsApp, Call Center 01 8000 123 456, sucursal) y pregunta en qué producto o trámite puedes ayudar.
3. NUNCA inventes tasas, costos, plazos ni fechas. NUNCA pidas OTP, claves ni datos sensibles.
4. "superCDT" = CDT Serfinanza (mismo producto).
5. Si hay PERFIL DEL CLIENTE, personaliza con lo que ya tiene. No ofrezcas lo que ya tiene.
6. Si hay SUGERENCIA PROACTIVA y encaja, inclúyela al final de forma natural.
7. Temas claramente ajenos al banco (deportes, política, recetas, etc.): redirige con cortesía hacia productos o trámites de Serfinanza.

ESTILO: 2-5 oraciones cortas; viñetas (-) solo si listas opciones. Segunda persona ("te", "tu").

CONTEXTO:
%s`

func (c *Client) GenerateForClient(ctx context.Context, req ClientRequest) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY no configurada")
	}

	system := fmt.Sprintf(clientSystemPrompt, buildContext(req.Chunks))
	if req.ProfileContext != "" {
		system += "\n\nPERFIL DEL CLIENTE:\n" + req.ProfileContext
	}
	if req.ProactiveHint != "" {
		system += "\n\nSUGERENCIA PROACTIVA (úsala solo si encaja):\n" + req.ProactiveHint
	}

	history := req.History
	if len(history) > 6 {
		history = history[len(history)-6:]
	}
	messages := make([]messageItem, 0, len(history)+1)
	for _, msg := range history {
		messages = append(messages, messageItem{Role: msg.Role, Content: msg.Content})
	}
	messages = append(messages, messageItem{Role: "user", Content: req.Question})

	body, _ := json.Marshal(messagesRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    system,
		Messages:  messages,
	})

	url := "https://api.anthropic.com/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(httpReq)
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

	return result.Content[0].Text, nil
}

func (c *Client) Generate(ctx context.Context, question, channel string, chunks []domain.Chunk) (string, error) {
	req := ClientRequest{
		Question: question,
		Chunks:   chunks,
	}
	return c.GenerateForClient(ctx, req)
}

func (c *Client) GenerateWithProfile(ctx context.Context, question, channel string, chunks []domain.Chunk, profileContext string) (string, error) {
	req := ClientRequest{
		Question:       question,
		Chunks:        chunks,
		ProfileContext: profileContext,
	}
	return c.GenerateForClient(ctx, req)
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
		return "(Sin fragmentos específicos para esta consulta; responde como asesor conversacional de Serfinanza e invita a concretar la necesidad.)"
	}

	var sb strings.Builder
	seen := make(map[string]bool)

	for _, c := range chunks {
		cleanDoc := cleanDocName(c.Doc)
		cleanSeccion := cleanSectionName(c.Seccion)
		cleanContenido := strings.TrimSpace(c.Contenido)

		if cleanContenido == "" || len(cleanContenido) < 20 {
			continue
		}

		key := cleanDoc + "|" + cleanSeccion + "|" + cleanContenido
		if seen[key] {
			continue
		}
		seen[key] = true

		if cleanSeccion != "" {
			sb.WriteString(cleanSeccion + ": " + cleanContenido + "\n\n")
		} else {
			sb.WriteString(cleanContenido + "\n\n")
		}
	}

	return sb.String()
}

func cleanDocName(doc string) string {
	doc = strings.ReplaceAll(doc, ".pdf", "")
	doc = strings.ReplaceAll(doc, ".txt", "")
	doc = strings.ReplaceAll(doc, ".md", "")
	doc = strings.ReplaceAll(doc, "_", " ")
	doc = strings.Title(doc)
	return doc
}

func cleanSectionName(seccion string) string {
	seccion = strings.ReplaceAll(seccion, "[PDF]", "")
	seccion = strings.ReplaceAll(seccion, "[TXT]", "")
	seccion = strings.ReplaceAll(seccion, "[MD]", "")
	seccion = strings.TrimSpace(seccion)
	return seccion
}
