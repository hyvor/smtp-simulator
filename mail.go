package main

import (
	"errors"
	"log"
	"net"
	"strings"

	netsmtp "net/smtp"

	"github.com/emersion/go-smtp"
)

type Mail struct {
	MailFrom string // Bounces, complaints are sent back to this address
	RcptTo   []string
	bounceActions map[string]Action
	complaints    []PendingComplaint
}

type PendingComplaint struct {
	OriginalMailFrom string
	To               string
	Delay            int
}

func NewMail() Mail {
	return Mail{
		MailFrom: "",
		RcptTo:   []string{},
		
		bounceActions: map[string]Action{},
		complaints:    []PendingComplaint{},
	}
}

var ErrorNoMailFrom = &smtp.SMTPError{
	Code:    503,
	Message: "No MAIL FROM specified",
}
var ErrorNoRcptTo = &smtp.SMTPError{
	Code:    503,
	Message: "No RCPT TO specified",
}

func (m *Mail) Rcpt(to string) error {

	local, _ := splitAddress(to)

	action, exists := localPartToAction[strings.ToLower(local)]

	if !exists {
		return &smtp.SMTPError{
			Code:    550,
			Message: "Unknown local part",
		}
	}

	if action.Type == ActionTypeSyncResponse {
		if action.Code == 250 {
			m.RcptTo = append(m.RcptTo, to) // add RCPT
			return nil
		}

		return &smtp.SMTPError{
			Code:         action.Code,
			EnhancedCode: action.EnhancedCode.Int(),
			Message:      action.Message,
		}
	} else if action.Type == ActionTypeAsyncBounce {
		m.bounceActions[to] = action
	} else if action.Type == ActionTypeAsyncComplaint {
		m.complaints = append(m.complaints, PendingComplaint{
			OriginalMailFrom: m.MailFrom,
			To:               to,
			Delay:            action.AsyncDelay,
		})
	}

	// reached for async actions
	m.RcptTo = append(m.RcptTo, to)

	return nil // OK Response
}

func (m *Mail) Complete() error {

	if m.MailFrom == "" {
		return ErrorNoMailFrom
	}

	if len(m.RcptTo) == 0 {
		return ErrorNoRcptTo
	}

	if len(m.bounceActions) > 0 {
		delay := 0
		for _, action := range m.bounceActions {
			if action.AsyncDelay > delay {
				delay = action.AsyncDelay
			}
		}

		go sendBounces(m.MailFrom, m.bounceActions, delay)
	}

	for _, complaint := range m.complaints {
		go sendComplaint(complaint.OriginalMailFrom, complaint.To, complaint.Delay)
	}

	return nil

}

func splitAddress(address string) (local, domain string) {
	parts := strings.Split(address, "@")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return address, ""
}

var smtpSendMail = netsmtp.SendMail
var sendMail = sendMailHandler
var netLookupMX = net.LookupMX

func sendMailHandler(to string, body string) error {

	_, domain := splitAddress(to)

	if domain == "" {
		return errors.New("invalid MAIL FROM address")
	}

	mxRecords, err := netLookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		log.Fatalf("Could not find MX records for %s: %v", domain, err)
	}

	mxHost := strings.TrimSuffix(mxRecords[0].Host, ".")

	log.Println("Sending email to", to, "via domain", mxHost)

	err = smtpSendMail(
		mxHost+":25",
		nil,
		"simulator@"+getDomain(),
		[]string{to},
		[]byte(body),
	)

	return err

}
