package smtpproxy

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/emersion/go-smtp"
)

// The Backend implements SMTP server methods.
type backend struct {
	userName string
	password string

	out chan<- *Message
}

// Login handles a login command with username and password.
func (bkd *backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	if username != bkd.userName || password != bkd.password {
		return nil, errors.New("Invalid username or password")
	}
	return &session{
		message: new(Message),
		out:     bkd.out,
	}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return nil, smtp.ErrAuthRequired
}

// A Session is returned after successful login.
type session struct {
	message *Message
	out     chan<- *Message
}

func (s *session) Mail(from string) error {
	s.message.From = from
	return nil
}

func (s *session) Rcpt(to string) error {
	s.message.To = to
	return nil
}

func (s *session) Data(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read data: %v", err)
	}

	s.message.Data = b
	s.out <- s.message

	return nil
}

func (s *session) Reset() {}

func (s *session) Logout() error {
	return nil
}
