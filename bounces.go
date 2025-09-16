package main

import (
	"bytes"
	"log"
	"text/template"
)

func sendBounces(originalMailFrom string, bounceActions map[string]Action) {

	data := DnsTemplateData{
		Subject:          "Delivery Status Notification (Failure)",
		To:               originalMailFrom,
		PlainTextMessage: "This is an automatically generated Delivery Status Notification.\n\nDelivery to the following recipients failed permanently:\n",
		Recipients:       []DsnRecipient{},
	}

	for addr, action := range bounceActions {
		data.PlainTextMessage += "- " + addr + ": " + action.Message + "\n"
		data.Recipients = append(data.Recipients, DsnRecipient{
			Address:        addr,
			EnhancedStatus: action.EnhancedCode.String(),
		})
	}

	emailBody, err := RenderDsnTemplate(data)

	if err != nil {
		log.Println("Error rendering DSN template:", err)
		return
	}

	err = sendMail(originalMailFrom, emailBody)

	if err != nil {
		log.Println("Error sending DSN email:", err)
		return
	}

}

// TEMPLATE ============

const DSN_TEMPLATE = `From: Hyvor SMTP Simulator <simulator@{{.Domain}}>
Message-Id: <123456789@{{.Domain}}>
Subject: {{.Subject}}
To: <{{.To}}>
MIME-Version: 1.0
Content-Type: multipart/report; report-type=delivery-status;
    boundary="__boundary__"

--__boundary__
Content-Type: text/plain;

{{.PlainTextMessage}}

--__boundary__
Content-Type: message/delivery-status

Reporting-MTA: dns; {{.Domain}}

{{range .Recipients}}Original-Recipient: rfc822;{{.Address}}
Final-Recipient: rfc822;{{.Address}}
Action: failed
Status: {{.EnhancedStatus}}{{end}}

--__boundary__
Content-Type: message/rfc822

[original message goes here]

--__boundary__--`

type DnsTemplateData struct {
	Subject          string
	To               string
	PlainTextMessage string
	Recipients       []DsnRecipient

	Domain string
}

type DsnRecipient struct {
	Address        string
	EnhancedStatus string
}

func RenderDsnTemplate(data DnsTemplateData) (string, error) {
	data.Domain = getDomain()

	tmpl, err := template.New("dsn").Parse(DSN_TEMPLATE)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
