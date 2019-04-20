package smtpproxy

import (
	"errors"
	"fmt"
	"time"

	"github.com/emersion/go-smtp"
)

const (
	DefaultListenPort  = 25
	DefaultDomain      = "localhost"
	DefaultReadTimeout = 5 * time.Second
	DefaultBufferLen   = 1024
)

type Message struct {
	From string
	To   string
	Data string
}

type Proxy struct {
	listenPort int

	userName    string
	password    string
	domain      string
	readTimeout time.Duration

	bufferLen int
}

type Option func(*Proxy) error

func ListenPort(port int) Option {
	return func(proxy *Proxy) error {
		proxy.listenPort = port
		return nil
	}
}

func Domain(domain string) Option {
	return func(proxy *Proxy) error {
		proxy.domain = domain
		return nil
	}
}

func ReadTimeout(timeout time.Duration) Option {
	return func(proxy *Proxy) error {
		proxy.readTimeout = timeout
		return nil
	}
}

func BufferLen(bufferLen int) Option {
	return func(proxy *Proxy) error {
		if bufferLen < 1 {
			return errors.New("invalid buffer length")
		}

		proxy.bufferLen = bufferLen
		return nil
	}
}

func NewProxy(userName, password string, opts ...Option) (*Proxy, error) {
	proxy := &Proxy{
		listenPort:  DefaultListenPort,
		userName:    userName,
		password:    password,
		domain:      DefaultDomain,
		readTimeout: DefaultReadTimeout,
		bufferLen:   DefaultBufferLen,
	}

	var err error
	for _, opt := range opts {
		err = opt(proxy)
		if err != nil {
			return nil, err
		}
	}

	return proxy, nil
}

func (proxy *Proxy) DoProxy() (<-chan *Message, <-chan error) {
	// make out buffer
	out := make(chan *Message, proxy.bufferLen)
	errCh := make(chan error, 1)

	// make backend
	be := &backend{
		userName: proxy.userName,
		password: proxy.password,
		out:      out,
	}

	// make smtp server
	s := smtp.NewServer(be)

	s.Addr = fmt.Sprintf(":%d", proxy.listenPort)
	s.Domain = proxy.domain
	s.ReadTimeout = proxy.readTimeout
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	go func() {
		errCh <- s.ListenAndServe()
	}()

	return out, errCh
}
