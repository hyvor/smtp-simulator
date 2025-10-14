package main

import (
	"io"
	"strings"
	"testing"

	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
)

func TestSmtpServer(t *testing.T) {

	smtpPort = "25251"

	s := NewSmtpServer()

	assert.NotNil(t, s)
	assert.Equal(t, "0.0.0.0:25251", s.Addr)
	assert.Equal(t, "localhost", s.Domain)

	sessionOriginal, err := s.Backend.NewSession(nil)
	assert.NoError(t, err)

	session, ok := sessionOriginal.(*Session)
	assert.True(t, ok)

	session.Mail("bounces@example.com", nil)
	assert.Equal(t, "bounces@example.com", session.mail.MailFrom)

	err = session.Rcpt("user@localhost", nil)
	assert.Equal(t, 0, len(session.mail.RcptTo)) // not added since wrong
	smtpErr, ok := err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 550, smtpErr.Code)
	assert.Equal(t, "Unknown local part", smtpErr.Message)

	err = session.Data(io.LimitReader(strings.NewReader("Test email body"), 1024))
	smtpErr, ok = err.(*smtp.SMTPError)
	assert.True(t, ok)
	assert.Equal(t, 503, smtpErr.Code)
	assert.Equal(t, "No RCPT TO specified", smtpErr.Message)

	session.Reset()
	assert.Empty(t, session.mail.RcptTo)

}
