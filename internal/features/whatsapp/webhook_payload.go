package whatsapp

import (
	"encoding/json"
	"strings"
)

type WebhookPayload struct {
	Event    string          `json:"event"`
	Instance string          `json:"instance"`
	Data     json.RawMessage `json:"data"`
}

type MessageData struct {
	Key         MessageKey `json:"key"`
	PushName    string     `json:"pushName"`
	Message     Message    `json:"message"`
	MessageType string     `json:"messageType"`
}

type MessageKey struct {
	RemoteJid string `json:"remoteJid"`
	FromMe    bool   `json:"fromMe"`
	ID        string `json:"id"`
}

type Message struct {
	Conversation string `json:"conversation"`
}

func ExtractIncomingText(payload WebhookPayload) (number, text string, ok bool) {
	event := strings.ToUpper(strings.ReplaceAll(payload.Event, ".", "_"))
	if event != "MESSAGES_UPSERT" {
		return "", "", false
	}

	var items []MessageData
	if err := json.Unmarshal(payload.Data, &items); err == nil && len(items) > 0 {
		return extractFromMessage(items[0])
	}

	var single MessageData
	if err := json.Unmarshal(payload.Data, &single); err != nil {
		return "", "", false
	}

	return extractFromMessage(single)
}

func extractFromMessage(msg MessageData) (number, text string, ok bool) {
	if msg.Key.FromMe {
		return "", "", false
	}

	number = normalizeJid(msg.Key.RemoteJid)
	text = strings.TrimSpace(msg.Message.Conversation)
	if number == "" || text == "" {
		return "", "", false
	}

	return number, text, true
}

func normalizeJid(jid string) string {
	jid = strings.TrimSpace(jid)
	if idx := strings.Index(jid, "@"); idx > 0 {
		jid = jid[:idx]
	}
	return jid
}
