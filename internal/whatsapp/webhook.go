package whatsapp

import (
	"encoding/json"
	"os"
	"strings"
)

type IncomingMessage struct {
	Phone string
	Text  string
	Event string
}

func AllowedNumber() string {
	n := os.Getenv("WHATSAPP_ALLOWED_NUMBER")
	if n == "" {
		n = os.Getenv("WHATSAPP_TARGET_NUMBER")
	}
	return NormalizePhone(n)
}

func UserProfileID() string {
	if id := os.Getenv("WHATSAPP_USER_PROFILE_ID"); id != "" {
		return id
	}
	return "C004"
}

func IsAllowedPhone(phone string) bool {
	allowed := AllowedNumber()
	if allowed == "" {
		return true
	}
	return NormalizePhone(phone) == allowed
}

func NormalizePhone(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "+")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")

	if at := strings.Index(s, "@"); at >= 0 {
		s = s[:at]
	}

	for len(s) > 0 && !isDigit(s[len(s)-1]) {
		s = s[:len(s)-1]
	}

	return s
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func ParseWebhook(raw []byte) (IncomingMessage, bool) {
	if msg, ok := parseSimple(raw); ok {
		return msg, true
	}
	if msg, ok := parseEvolution(raw); ok {
		return msg, true
	}
	return IncomingMessage{}, false
}

func parseSimple(raw []byte) (IncomingMessage, bool) {
	var payload struct {
		Event   string `json:"event"`
		Phone   string `json:"phone"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return IncomingMessage{}, false
	}
	if payload.Phone == "" || payload.Message == "" {
		return IncomingMessage{}, false
	}
	return IncomingMessage{
		Phone: NormalizePhone(payload.Phone),
		Text:  strings.TrimSpace(payload.Message),
		Event: payload.Event,
	}, true
}

func parseEvolution(raw []byte) (IncomingMessage, bool) {
	var envelope struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
		Body  *struct {
			Event string          `json:"event"`
			Data  json.RawMessage `json:"data"`
		} `json:"body"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return IncomingMessage{}, false
	}

	event := strings.ToLower(envelope.Event)
	data := envelope.Data
	if envelope.Body != nil {
		if envelope.Body.Event != "" {
			event = strings.ToLower(envelope.Body.Event)
		}
		if len(envelope.Body.Data) > 0 {
			data = envelope.Body.Data
		}
	}

	if event != "" && !isMessageUpsert(event) {
		return IncomingMessage{}, false
	}
	if len(data) == 0 {
		return IncomingMessage{}, false
	}

	if data[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(data, &items); err != nil || len(items) == 0 {
			return IncomingMessage{}, false
		}
		data = items[0]
	}

	return parseEvolutionData(event, data)
}

func isMessageUpsert(event string) bool {
	event = strings.ToLower(event)
	return event == "messages.upsert" || event == "messages_upsert"
}

type evolutionData struct {
	Key struct {
		RemoteJid string `json:"remoteJid"`
		FromMe    bool   `json:"fromMe"`
		SenderPn  string `json:"senderPn"`
	} `json:"key"`
	Message struct {
		Conversation        string `json:"conversation"`
		ExtendedTextMessage struct {
			Text string `json:"text"`
		} `json:"extendedTextMessage"`
	} `json:"message"`
}

func parseEvolutionData(event string, data json.RawMessage) (IncomingMessage, bool) {
	var payload evolutionData
	if err := json.Unmarshal(data, &payload); err != nil {
		return IncomingMessage{}, false
	}

	if payload.Key.FromMe {
		return IncomingMessage{}, false
	}

	text := strings.TrimSpace(payload.Message.Conversation)
	if text == "" {
		text = strings.TrimSpace(payload.Message.ExtendedTextMessage.Text)
	}
	if text == "" {
		return IncomingMessage{}, false
	}

	phone := phoneFromJID(payload.Key.RemoteJid, payload.Key.SenderPn)
	if phone == "" {
		return IncomingMessage{}, false
	}

	return IncomingMessage{
		Phone: phone,
		Text:  text,
		Event: event,
	}, true
}

func phoneFromJID(remoteJid, senderPn string) string {
	jid := strings.TrimSpace(remoteJid)
	if strings.HasSuffix(jid, "@lid") && senderPn != "" {
		jid = senderPn
	}
	return NormalizePhone(jid)
}
