package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderDsnTempl(t *testing.T) {

	data := DnsTemplateData{
		Subject:          "Delivery Status Notification (Failure)",
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

	fmt.Println(result)

	assert.Contains(t, result, "Subject: Delivery Status Notification (Failure)")
	assert.Contains(t, result, "To: <user@example.com>")
	assert.Contains(t, result, "Bad news everyone")

}
