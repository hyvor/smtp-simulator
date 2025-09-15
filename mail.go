package main

import (
	"log"
	"strings"

	"github.com/emersion/go-smtp"
)

type Mail struct {
	MailFrom string // Bounces, complaints are sent back to this address
	RcptTo   []string
}

var ErrorNoMailFrom = &smtp.SMTPError{
	Code:    503,
	Message: "No MAIL FROM specified",
}
var ErrorNoRcptTo = &smtp.SMTPError{
	Code:    503,
	Message: "No RCPT TO specified",
}

func (m *Mail) Process() *smtp.SMTPError {

	if m.MailFrom == "" {
		return ErrorNoMailFrom
	}

	if len(m.RcptTo) == 0 {
		return ErrorNoRcptTo
	}

	for _, to := range m.RcptTo {
		err := m.handleRcpt(to)
		if err != nil {
			return err
		}
	}

	return nil

}

func (m *Mail) handleRcpt(to string) *smtp.SMTPError {

	local, _ := splitAddress(to)
	log.Println("Processing for:", local)

	return nil

}

func splitAddress(address string) (local, domain string) {
	parts := strings.Split(address, "@")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return address, ""
}
