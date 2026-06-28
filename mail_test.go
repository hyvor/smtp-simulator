package main

import (
	"encoding/hex"
	"errors"
	"net"
	netsmtp "net/smtp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
)

// customAddr builds a custom+<hex>@localhost recipient from a raw JSON spec.
func customAddr(t *testing.T, spec string) string {
	t.Helper()
	return "custom+" + hex.EncodeToString([]byte(spec)) + "@localhost"
}

func TestProcessMailAsyncBounce(t *testing.T) {

	mail := NewMail()

	err := mail.Complete()
	assert.Equal(t, ErrorNoMailFrom, err)

	mail.MailFrom = "bounces@example.com"

	err = mail.Complete()
	assert.Equal(t, ErrorNoRcptTo, err)

	mail.Rcpt("missing+async@localhost")

	bouncesTo := ""
	bouncesActions := map[string]Action{}
	bouncesDelay := time.Duration(0)

	writeMx := sync.Mutex{}

	sendBounces = func(originalMailFrom string, bounceActions map[string]Action, delay time.Duration) {
		writeMx.Lock()
		defer writeMx.Unlock()
		bouncesTo = originalMailFrom
		bouncesActions = bounceActions
		bouncesDelay = delay
	}

	err = mail.Complete()

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

	assert.Equal(t, 250*time.Millisecond, bouncesDelay)

	mail.MailFrom = "bounce@example.com"
	err = mail.Rcpt("unknown@localhost")

	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "Unknown local part", smtpErr.Message)

}

func TestProcessMailImmediateBounce(t *testing.T) {

	mail := NewMail()

	mail.MailFrom = "bounces@example.com"
	err := mail.Rcpt("missing@localhost")
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)

	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "User unknown", smtpErr.Message)

}

func TestProcessMailComplaint(t *testing.T) {

	mail := NewMail()

	mail.MailFrom = "bounces@example.com"
	mail.Rcpt("complaint@localhost")

	complaintSendTo := ""
	complaintRecipient := ""
	complaintDelay := time.Duration(0)

	mx := sync.Mutex{}

	sendComplaint = func(originalMailFrom string, to string, delay time.Duration) {
		mx.Lock()
		defer mx.Unlock()
		complaintSendTo = originalMailFrom
		complaintRecipient = to
		complaintDelay = delay
	}

	// Simulate processing the mail
	err := mail.Complete()

	time.Sleep(20 * time.Millisecond) // Wait for goroutine to finish

	mx.Lock()
	defer mx.Unlock()

	assert.NoError(t, err)
	assert.Equal(t, "bounces@example.com", complaintSendTo)
	assert.Equal(t, "complaint@localhost", complaintRecipient)
	assert.Equal(t, 250*time.Millisecond, complaintDelay)
}

func TestMailWithOneRcptOkayOtherFail(t *testing.T) {

	mail := NewMail()

	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt("accept@localhost")
	assert.Nil(t, err)

	err = mail.Rcpt("busy@localhost")
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 450, smtpErr.Code)

	assert.Equal(t, []string{"accept@localhost"}, mail.RcptTo)

	err = mail.Complete()
	assert.Nil(t, err)

}

func TestMailWithOneRcptOkayOneFailAndOtherAsync(t *testing.T) {

	mail := NewMail()

	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt("accept@localhost")
	assert.Nil(t, err)

	err = mail.Rcpt("busy@localhost")
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 450, smtpErr.Code)

	err = mail.Rcpt("missing+async@localhost")
	assert.Nil(t, err)

	assert.Equal(t, []string{"accept@localhost", "missing+async@localhost"}, mail.RcptTo)

	err = mail.Complete()
	assert.Nil(t, err)

}

func TestMailAtData(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	// RCPT should accept with 250 OK (no error)
	err := mail.Rcpt("busy+atdata@localhost")
	assert.Nil(t, err)
	assert.Equal(t, []string{"busy+atdata@localhost"}, mail.RcptTo)

	// DATA should return the error
	err = mail.Complete()
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 450, smtpErr.Code)
	assert.Equal(t, "Requested mail action not taken: mailbox busy", smtpErr.Message)

}

func TestCustomActionAtRcpt(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt(customAddr(t, `{"status":550,"message":"5.7.1 bad reputation"}`))

	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "5.7.1 bad reputation", smtpErr.Message)
	assert.Equal(t, smtp.NoEnhancedCode, smtpErr.EnhancedCode)
	assert.Empty(t, mail.RcptTo)
}

func TestCustomActionAtData(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	addr := customAddr(t, `{"status":550,"message":"5.7.1 bad reputation","at":"data"}`)

	// RCPT accepts with 250 OK (no error)
	err := mail.Rcpt(addr)
	assert.Nil(t, err)
	assert.Equal(t, []string{addr}, mail.RcptTo)

	// DATA returns the error
	err = mail.Complete()
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "5.7.1 bad reputation", smtpErr.Message)
	assert.Equal(t, smtp.NoEnhancedCode, smtpErr.EnhancedCode)
}

func TestCustomActionAtOmittedDefaultsToRcpt(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt(customAddr(t, `{"status":450,"message":"slow down"}`))

	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 450, smtpErr.Code)
	assert.Empty(t, mail.RcptTo)
}

func TestCustomActionVerbatimMessage(t *testing.T) {

	// Message is emitted verbatim; EnhancedCode is NoEnhancedCode so the caller
	// controls whether an enhanced code appears at all.
	action, err := parseCustomAction("custom+" + hex.EncodeToString([]byte(`{"status":550,"message":"bad reputation"}`)))
	assert.NoError(t, err)
	assert.Equal(t, EnhancedCode{-1, -1, -1}, action.EnhancedCode)
	assert.Equal(t, "bad reputation", action.Message)
	assert.Equal(t, 550, action.Code)
	assert.Equal(t, ActionTypeSyncResponse, action.Type)
}

func TestCustomActionAccept(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	addr := customAddr(t, `{"status":250,"message":"OK"}`)

	err := mail.Rcpt(addr)
	assert.Nil(t, err)
	assert.Equal(t, []string{addr}, mail.RcptTo)

	err = mail.Complete()
	assert.Nil(t, err)
}

func TestCustomActionInvalidHex(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt("custom+zzzz@localhost")
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 501, smtpErr.Code)
}

func TestCustomActionInvalidJSON(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt(customAddr(t, "not json"))
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 501, smtpErr.Code)
}

func TestCustomActionInvalidAt(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt(customAddr(t, `{"status":550,"message":"x","at":"helo"}`))
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 501, smtpErr.Code)
	assert.Contains(t, smtpErr.Message, "at")
}

func TestCustomActionInvalidStatus(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	for _, spec := range []string{`{"status":0,"message":"x"}`, `{"status":700,"message":"x"}`} {
		err := mail.Rcpt(customAddr(t, spec))
		smtpErr, ok := err.(*smtp.SMTPError)
		assert.True(t, ok, spec)
		assert.Equal(t, 501, smtpErr.Code, spec)
	}
}

func TestCustomActionAtDataWith2xxRejected(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	err := mail.Rcpt(customAddr(t, `{"status":250,"message":"OK","at":"data"}`))
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 501, smtpErr.Code)
	assert.Contains(t, smtpErr.Message, "2xx")
	assert.Empty(t, mail.RcptTo)
}

func TestCustomActionUppercaseHex(t *testing.T) {

	mail := NewMail()
	mail.MailFrom = "bounces@example.com"

	encoded := strings.ToUpper(hex.EncodeToString([]byte(`{"status":550,"message":"5.7.1 bad reputation"}`)))
	err := mail.Rcpt("custom+" + encoded + "@localhost")

	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "5.7.1 bad reputation", smtpErr.Message)
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

func TestSendMailHandlerARecord(t *testing.T) {

	netLookupMX = func(domain string) ([]*net.MX, error) {
		return []*net.MX{}, nil
	}

	netLookupHost = func(domain string) ([]string, error) {
		return []string{"192.0.2.1"}, nil
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

	assert.Equal(t, "example.com:25", sendToAddr)
	assert.Equal(t, "simulator@localhost", fromGot)
	assert.Equal(t, []string{"bounces@example.com"}, toGot)

}
