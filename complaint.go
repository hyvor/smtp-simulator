package main

import (
	"bytes"
	"log"
	"text/template"
	"time"
)

var sendComplaint = sendComplaintHandler

func sendComplaintHandler(originalMailFrom string, to string, delay int) {
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Second)
	}

	data := ArfTemplateData{
		To:                originalMailFrom,
		OriginalRecipient: to,

		Domain: getDomain(),
	}

	emailBody, err := RenderArfTemplate(data)

	if err != nil {
		log.Println("Error rendering ARF template:", err)
		return
	}

	err = sendMail(originalMailFrom, emailBody)

	if err != nil {
		log.Println("Error sending ARF email:", err)
		return
	}

}

// TEMPLATE ============

const ARF_TEMPLATE = `From: <abuse@{{.Domain}}>
Subject: Abuse Report
To: <{{.To}}>
MIME-Version: 1.0
Content-Type: multipart/report; report-type=feedback-report;
     boundary="__boundary__"

--__boundary__
Content-Type: text/plain; charset="US-ASCII"
Content-Transfer-Encoding: 7bit

This is an email abuse report for an email message received from IP
192.0.2.2 on Thu, 8 Sept 2025 14:00:00 EDT.

--__boundary__
Content-Type: message/feedback-report

Feedback-Type: abuse
Original-Mail-From: <{{.To}}>
Original-Rcpt-To: <{{.OriginalRecipient}}>
Received-Date: Thu, 8 Sept 2025 14:00:00 EDT
Source-IP: 192.0.2.2

--__boundary__
Content-Type: message/rfc822
Content-Disposition: inline

Subject: Earn money fast!

Spam Spam Spam
Spam Spam Spam
Spam Spam Spam
Spam Spam Spam
--__boundary__--`

type ArfTemplateData struct {
	To                string // original MAIL FROM
	OriginalRecipient string // compliant@...

	Domain string
}

func RenderArfTemplate(data ArfTemplateData) (string, error) {
	data.Domain = getDomain()

	tmpl, err := template.New("arf").Parse(ARF_TEMPLATE)
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
