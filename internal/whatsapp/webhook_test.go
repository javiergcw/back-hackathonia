package whatsapp

import "testing"

func TestNormalizePhone(t *testing.T) {
	cases := map[string]string{
		"573168731521":                    "573168731521",
		"+57 316 873 1521":                "573168731521",
		"573168731521@s.whatsapp.net":     "573168731521",
		"573168731521@lid":                "573168731521",
	}
	for input, want := range cases {
		if got := NormalizePhone(input); got != want {
			t.Fatalf("NormalizePhone(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseSimpleWebhook(t *testing.T) {
	raw := []byte(`{
		"event": "MESSAGES_UPSERT",
		"phone": "573168731521",
		"message": "¿Qué es un CDT?"
	}`)
	msg, ok := ParseWebhook(raw)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if msg.Phone != "573168731521" || msg.Text != "¿Qué es un CDT?" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestParseEvolutionWebhook(t *testing.T) {
	raw := []byte(`{
		"event": "messages.upsert",
		"instance": "javierg",
		"data": {
			"key": {
				"remoteJid": "573168731521@s.whatsapp.net",
				"fromMe": false
			},
			"message": {
				"conversation": "Hola Serfinanza"
			}
		}
	}`)
	msg, ok := ParseWebhook(raw)
	if !ok {
		t.Fatal("expected parse ok")
	}
	if msg.Phone != "573168731521" || msg.Text != "Hola Serfinanza" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestParseEvolutionIgnoresFromMe(t *testing.T) {
	raw := []byte(`{
		"event": "messages.upsert",
		"data": {
			"key": {
				"remoteJid": "573168731521@s.whatsapp.net",
				"fromMe": true
			},
			"message": {
				"conversation": "respuesta del bot"
			}
		}
	}`)
	if _, ok := ParseWebhook(raw); ok {
		t.Fatal("expected fromMe message to be ignored")
	}
}

func TestIsAllowedPhone(t *testing.T) {
	t.Setenv("WHATSAPP_ALLOWED_NUMBER", "573168731521")

	if !IsAllowedPhone("573168731521@s.whatsapp.net") {
		t.Fatal("expected allowed phone")
	}
	if IsAllowedPhone("573024158002") {
		t.Fatal("expected other phone to be rejected")
	}
}
