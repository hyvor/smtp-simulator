package main

import (
	"errors"
	"net"
	netsmtp "net/smtp"
	"sync"
	"testing"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
)

func TestProcessMailAsyncBounce(t *testing.T) {

	mail := NewMail()

	err := mail.Process()
	assert.Equal(t, ErrorNoMailFrom, err)

	mail.MailFrom = "bounces@example.com"

	err = mail.Process()
	assert.Equal(t, ErrorNoRcptTo, err)

	mail.RcptTo = []string{"missing+async@localhost"}

	bouncesTo := ""
	bouncesActions := map[string]Action{}
	bouncesDelay := 1

	writeMx := sync.Mutex{}

	sendBounces = func(originalMailFrom string, bounceActions map[string]Action, delaySeconds int) {
		writeMx.Lock()
		defer writeMx.Unlock()
		bouncesTo = originalMailFrom
		bouncesActions = bounceActions
		bouncesDelay = delaySeconds
	}

	err = mail.Process()

	time.Sleep(20 * time.Millisecond) // Wait for goroutine to finish

	assert.NoError(t, err)
	writeMx.Lock()
	defer writeMx.Unlock()
	assert.Equal(t, "bounces@example.com", bouncesTo)
	_, exists := bouncesActions["missing+async@localhost"]
	assert.True(t, exists)
	action := bouncesActions["missing+async@localhost"]
	assert.Equal(t, "User unknown", action.Message)
	assert.Equal(t, "5.1.1", action.EnhancedCode.String())
	assert.Equal(t, 550, action.Code)

	assert.Equal(t, 0, bouncesDelay)

	mail.MailFrom = "bounce@example.com"
	mail.RcptTo = []string{"unknown@localhost"}

	err = mail.Process()

	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "Unknown local part", smtpErr.Message)

}

func TestProcessMailImmediateBounce(t *testing.T) {

	mail := NewMail()

	mail.MailFrom = "bounces@example.com"
	mail.RcptTo = []string{"missing@localhost"}

	err := mail.Process()

	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "User unknown", smtpErr.Message)

}

func TestProcessMailComplaint(t *testing.T) {

	mail := NewMail()

	mail.MailFrom = "bounces@example.com"
	mail.RcptTo = []string{"complaint@localhost"}

	complaintSendTo := ""
	complaintRecipient := ""
	complaintDelay := 1

	mx := sync.Mutex{}

	sendComplaint = func(originalMailFrom string, to string, delay int) {
		mx.Lock()
		defer mx.Unlock()
		complaintSendTo = originalMailFrom
		complaintRecipient = to
		complaintDelay = delay
	}

	// Simulate processing the mail
	err := mail.Process()

	time.Sleep(20 * time.Millisecond) // Wait for goroutine to finish

	mx.Lock()
	defer mx.Unlock()

	assert.NoError(t, err)
	assert.Equal(t, "bounces@example.com", complaintSendTo)
	assert.Equal(t, "complaint@localhost", complaintRecipient)
	assert.Equal(t, 0, complaintDelay)
}

func TestSplitAddress(t *testing.T) {

	tests := []struct {
		address        string
		expectedLocal  string
		expectedDomain string
	}{
		{"test@hyvor.com", "test", "hyvor.com"},
		{"supun+contact@gmail.com", "supun+contact", "gmail.com"},
		{"invalidaddress", "invalidaddress", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		local, domain := splitAddress(tt.address)
		assert.Equal(t, tt.expectedLocal, local)
		assert.Equal(t, tt.expectedDomain, domain)
	}

}

func TestSendMailhandler(t *testing.T) {

	netLookupMX = func(domain string) ([]*net.MX, error) {
		if domain == "example.com" {
			return []*net.MX{
				{Host: "mx1.example.com.", Pref: 10},
				{Host: "mx2.example.com.", Pref: 20},
			}, nil
		}
		return nil, errors.New("no MX records found")
	}

	sendToAddr := ""
	fromGot := ""
	toGot := []string{}

	smtpSendMail = func(addr string, _ netsmtp.Auth, from string, to []string, msg []byte) error {
		sendToAddr = addr
		fromGot = from
		toGot = to
		return nil
	}

	err := sendMailHandler("bounces@example.com", "Test email body")
	assert.NoError(t, err)

	assert.Equal(t, "mx1.example.com:25", sendToAddr)
	assert.Equal(t, "simulator@localhost", fromGot)
	assert.Equal(t, []string{"bounces@example.com"}, toGot)

}
