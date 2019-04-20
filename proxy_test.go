package smtpproxy

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"testing"
	"time"
)

const (
	testUserName = "scryner"
	testPassword = "pass"
	testPort     = 9925
)

func TestProxy(t *testing.T) {
	proxy, err := NewProxy(testUserName, testPassword, ListenPort(testPort))
	if err != nil {
		t.Error("failed to make proxy:", err)
		t.FailNow()
	}

	out, errCh := proxy.DoProxy()
	time.Sleep(time.Second)

	cliErrCh := make(chan error)

	go func() {
		// setup smtp request
		auth := smtp.PlainAuth("", testUserName, testPassword, "localhost")

		to := []string{"recipient@example.net"}
		msg := []byte("To: recipient@example.net\r\n" +
			"Subject: discount Gophers!\r\n" +
			"\r\n" +
			"This is the email body.\r\n")

		// send email
		addr := fmt.Sprintf("localhost:%d", testPort)
		err := smtp.SendMail(addr, auth, "sender@example.org", to, msg)

		if err != nil {
			cliErrCh <- err
		}
	}()

	for {
		select {
		case msg := <-out:
			b, _ := json.MarshalIndent(msg, "", "  ")
			t.Log("message =>", string(b))
			return

		case err := <-errCh:
			t.Error("failed to listen server:", err)
			t.FailNow()
		case err := <-cliErrCh:
			t.Error("failed to send mail:", err)
			t.FailNow()
		}
	}
}
