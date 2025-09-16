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
}

func NewMail() Mail {
	return Mail{
		MailFrom: "",
		RcptTo:   []string{},
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

func (m *Mail) Process() error {

	if m.MailFrom == "" {
		return ErrorNoMailFrom
	}

	if len(m.RcptTo) == 0 {
		return ErrorNoRcptTo
	}

	bounceActions := map[string]Action{}

	for _, to := range m.RcptTo {
		err := m.handleRcpt(to, &bounceActions)
		if err != nil {
			return err
		}
	}

	if len(bounceActions) > 0 {
		go sendBounces(m.MailFrom, bounceActions)
	}

	return nil

}

func (m *Mail) handleRcpt(to string, bounceActions *map[string]Action) *smtp.SMTPError {

	local, _ := splitAddress(to)

	action, exists := localPartToAction[strings.ToLower(local)]

	if !exists {
		return &smtp.SMTPError{
			Code:    550,
			Message: "Unknown local part",
		}
	}

	if action.Type == ActionTypeSyncResponse {
		// note: SMTPError works for 200 as well
		return &smtp.SMTPError{
			Code:         action.Code,
			EnhancedCode: action.EnhancedCode.Int(),
			Message:      action.Message,
		}
	} else if action.Type == ActionTypeAsyncBounce {
		(*bounceActions)[to] = action
	} else if action.Type == ActionTypeAsyncComplaint {
		go sendComplaint(m.MailFrom, to, action.AsyncDelay)
	}

	return nil // OK response

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

func sendMailHandler(to string, body string) error {

	_, domain := splitAddress(to)

	if domain == "" {
		return errors.New("invalid MAIL FROM address")
	}

	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		log.Fatalf("Could not find MX records for %s: %v", domain, err)
	}

	mxHost := strings.TrimSuffix(mxRecords[0].Host, ".")

	log.Println("Sending email to", to, "via domain", mxHost)

	err = smtpSendMail(
		mxHost+":25",
		nil,
		"from@hyvor-smtp-simulator.com",
		[]string{to},
		[]byte(body),
	)

	return err

}
