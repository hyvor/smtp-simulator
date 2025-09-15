package main

import (
	"io"
	"time"

	"github.com/emersion/go-smtp"
)

type Backend struct{}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		mail: Mail{},
	}, nil
}

type Session struct {
	mail Mail
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.mail.MailFrom = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.mail.RcptTo = append(s.mail.RcptTo, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	s.mail.Process()
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

func NewSmtpServer() *smtp.Server {

	be := &Backend{}
	s := smtp.NewServer(be)

	s.Addr = "localhost:1025"
	s.Domain = "localhost"
	s.WriteTimeout = 10 * time.Second
	s.ReadTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	return s
}
