// internal/email/smtp.go
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
)

// SMTPSender implements Sender using SMTP
type SMTPSender struct {
	host      string
	port      string
	username  string
	password  string
	from      string
	tlsConfig *tls.Config
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		tlsConfig: &tls.Config{
			ServerName: host,
		},
	}
}

func (s *SMTPSender) Send(ctx context.Context, to, subject, body string, isHTML bool) error {
	// Set content type
	contentType := "text/plain; charset=UTF-8"
	if isHTML {
		contentType = "text/html; charset=UTF-8"
	}

	// Build message
	msg := fmt.Appendf(nil,
		"To: %s\r\n"+
			"From: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: %s\r\n"+
			"\r\n"+
			"%s\r\n",
		to, s.from, subject, contentType, body,
	)

	// Connect with timeout
	addr := net.JoinHostPort(s.host, s.port)

	// Use a goroutine to handle context cancellation
	done := make(chan error, 1)
	go func() {
		// For production, use STARTTLS or TLS
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		done <- smtp.SendMail(addr, auth, s.from, []string{to}, msg)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// MockSender for development/testing (logs instead of sending)
type MockSender struct{}

func NewMockSender() *MockSender {
	return &MockSender{}
}

func (m *MockSender) Send(ctx context.Context, to, subject, body string, isHTML bool) error {
	fmt.Printf("\n📧 MOCK EMAIL\nTo: %s\nSubject: %s\nBody: %s\n\n", to, subject, body)
	return nil
}
