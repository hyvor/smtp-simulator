package main

import (
	"bytes"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendComplaint(t *testing.T) {

	sentTo := ""
	sentBody := ""

	sendMail = func(to, body string) error {
		sentTo = to
		sentBody = body
		return nil
	}

	sendComplaint("sender@example.com", "complaint@localhost", 0)

	assert.Equal(t, "sender@example.com", sentTo)
	assert.Contains(t, sentBody, "Subject: Abuse Report")
	assert.Contains(t, sentBody, "This is an email abuse report for an email message received from IP")
	assert.Contains(t, sentBody, "Original-Mail-From: <sender@example.com>")
	assert.Contains(t, sentBody, "Original-Rcpt-To: <complaint@localhost>")

}

func TestRenderArfTemplate(t *testing.T) {

	data := ArfTemplateData{
		To:                "sender@example.com",
		OriginalRecipient: "complaint@localhost",
	}

	result, err := RenderArfTemplate(data)
	assert.NoError(t, err)

	assert.Contains(t, result, "Subject: Abuse Report")
	assert.Contains(t, result, "To: <sender@example.com>")
	assert.Contains(t, result, "Original-Mail-From: <sender@example.com>")
	assert.Contains(t, result, "Original-Rcpt-To: <complaint@localhost>")
	assert.Contains(t, result, "Feedback-Type: abuse")
	assert.Contains(t, result, "Spam Spam Spam")

	msg, err := mail.ReadMessage(bytes.NewReader([]byte(result)))
	assert.NoError(t, err)
	assert.Equal(t, "Abuse Report", msg.Header.Get("Subject"))
	assert.Contains(t, msg.Header.Get("From"), "abuse@localhost")

}
