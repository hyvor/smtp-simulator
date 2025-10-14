package main

import (
	"io"
	"log"
	"time"

	"github.com/emersion/go-smtp"
)

type Backend struct{}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		mail: NewMail(),
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
	if s.mail.MailFrom == "" {
		return ErrorNoMailFrom
	}

	return s.mail.Rcpt(to)
}

func (s *Session) Data(r io.Reader) error {
	_, err := io.ReadAll(r)

	if err != nil {
		log.Println("Error reading email body:", err)
		return err
	}

	return s.mail.Complete()
}

func (s *Session) Reset() {
	s.mail = NewMail()
}

func (s *Session) Logout() error {
	return nil
}

var smtpPort = "25"

func NewSmtpServer() *smtp.Server {

	be := &Backend{}
	s := smtp.NewServer(be)

	s.Addr = "0.0.0.0:" + smtpPort
	s.Domain = getDomain()
	s.WriteTimeout = 10 * time.Second
	s.ReadTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	return s
}
