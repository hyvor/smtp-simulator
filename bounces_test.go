package main

import (
	"bytes"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendBounces(t *testing.T) {

	sentTo := ""
	sentBody := ""

	sendMail = func(to, body string) error {
		sentTo = to
		sentBody = body
		return nil
	}

	sendBounces("bounces@hyvor.com", map[string]Action{
		"failed@example.com": {
			Message:      "User not found",
			EnhancedCode: EnhancedCode([3]int{5, 1, 1}),
		},
	}, 0)

	assert.Equal(t, "bounces@hyvor.com", sentTo)
	assert.Contains(t, sentBody, "This is an automatically generated Delivery Status Notification.")
	assert.Contains(t, sentBody, "Delivery to the following recipients failed permanently:")
	assert.Contains(t, sentBody, "- failed@example.com: User not found")

}

func TestRenderDsnTempl(t *testing.T) {

	data := DnsTemplateData{
		To:               "user@example.com",
		PlainTextMessage: "Bad news everyone",
		Recipients: []DsnRecipient{
			{
				Address:        "user@example.com",
				EnhancedStatus: "5.1.1",
			},
		},
	}

	result, err := RenderDsnTemplate(data)
	assert.NoError(t, err)

	assert.Contains(t, result, "Subject: Delivery Status Notification (Failure)")
	assert.Contains(t, result, "To: <user@example.com>")
	assert.Contains(t, result, "Bad news everyone")

	msg, err := mail.ReadMessage(bytes.NewReader([]byte(result)))
	assert.NoError(t, err)
	assert.Equal(t, "Delivery Status Notification (Failure)", msg.Header.Get("Subject"))
	assert.Equal(t, "Hyvor SMTP Simulator <simulator@localhost>", msg.Header.Get("From"))

}
