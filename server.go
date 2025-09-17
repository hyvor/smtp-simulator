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
	s.mail.RcptTo = append(s.mail.RcptTo, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	_, err := io.ReadAll(r)

	if err != nil {
		log.Println("Error reading email body:", err)
		return err
	}

	err = s.mail.Process()

	return err
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

func NewSmtpServer() *smtp.Server {

	be := &Backend{}
	s := smtp.NewServer(be)

	s.Addr = "0.0.0.0:25"
	s.Domain = "localhost"
	s.WriteTimeout = 50 * time.Second
	s.ReadTimeout = 50 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	return s
}
