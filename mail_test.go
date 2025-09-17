package main

import (
	"errors"
	"net"
	netsmtp "net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
