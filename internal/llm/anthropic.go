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

const clientSystemPrompt = `Eres el asesor virtual de Banco Serfinanza. Hablas directamente con un CLIENTE.

Tu misión: orientarlo con claridad, calidez y confianza, como un buen asesor del banco que lo conoce.

REGLAS:
1. El CONTEXTO es la documentación oficial vigente de Serfinanza. Preséntala con seguridad: afirma los datos, no los supongas ni los cuestiones.
2. NO digas "según el PDF", "según la información oficial", nombres de archivos ni secciones.
3. NO uses frases dubitativas ("podría", "quizás", "creo que"). Si está en el contexto, dilo con convicción.
4. Si la información NO está en el contexto, dilo con honestidad y sugiere App, Web, WhatsApp, Call Center o Sucursal según corresponda.
5. NUNCA inventes tasas, costos, plazos ni fechas. NUNCA pidas OTP, claves ni datos sensibles.
6. Tono: cercano, profesional, en segunda persona ("te", "tu"). Trata al usuario como cliente, no como empleado del banco.
7. "superCDT" = CDT Serfinanza (mismo producto).
8. Estilo: Responde con claridad y detalle. Usa 2-4 oraciones cortas. Si necesitas listar elementos, puedes usar viñetas simples (-). Máximo 5-6 oraciones en total.
9. Si hay PERFIL DEL CLIENTE, personaliza la respuesta con lo que ya tiene contratado. No le ofrezcas lo que ya tiene.
10. Si hay SUGERENCIA PROACTIVA y encaja con la pregunta, inclúyela al final de forma natural.
11. ALCANCE: Solo respondes sobre Banco Serfinanza (productos, trámites, App, tarjetas, CDT, créditos, pagos, extractos, seguridad). Si la pregunta no es del banco, responde EXACTAMENTE: "Solo puedo ayudarte con productos, trámites y servicios de Banco Serfinanza. ¿Tienes alguna consulta sobre tu cuenta, tarjeta, CDT, crédito o la App?"
12. NUNCA respondas preguntas ajenas al banco aunque sepas la respuesta.

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
	if len(history) > 4 {
		history = history[len(history)-4:]
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
		return ""
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
