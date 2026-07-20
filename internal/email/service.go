// internal/email/service.go
package email

import (
	"context"
	"fmt"
	"log"
	"sync"

	"golangchatapp/internal/events"
)

// EmailPayload is the data needed to send an email
type EmailPayload struct {
	To      string            `json:"to"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	HTML    bool              `json:"html"`
	Meta    map[string]string `json:"meta,omitempty"` // Template vars, etc.
}

// Service handles email operations
type Service struct {
	bus       *events.EventBus
	sender    Sender
	templates map[string]string // Template cache
	wg        sync.WaitGroup    // add this
}

// Sender interface for different providers (SMTP, SendGrid, AWS SES, etc.)
type Sender interface {
	Send(ctx context.Context, to, subject, body string, isHTML bool) error
}

// NewService creates an email service that listens for events
func NewService(bus *events.EventBus, sender Sender) *Service {
	s := &Service{
		bus:       bus,
		sender:    sender,
		templates: make(map[string]string),
	}
	return s
}

// Start begins listening for email events in a goroutine
// func (s *Service) Start(ctx context.Context) {
// 	ch := s.bus.Subscribe(
// 		events.EventUserRegistered,
// 		events.EventPasswordReset,
// 		events.EventMessageReceived,
// 	)
//
// 	go func() {
// 		defer s.bus.Unsubscribe(ch)
//
// 		for {
// 			select {
// 			case event, ok := <-ch:
// 				if !ok {
// 					log.Println("Email service: event channel closed")
// 					return
// 				}
// 				s.handleEvent(ctx, event)
//
// 			case <-ctx.Done():
// 				log.Println("Email service: shutting down")
// 				return
// 			}
// 		}
// 	}()
// }
//

// // internal/email/service.go — Start() method
// func (s *Service) Start(ctx context.Context) {
// 	ch := s.bus.Subscribe(
// 		events.EventUserRegistered,
// 		events.EventPasswordReset,
// 		events.EventMessageReceived,
// 	)
//
// 	go func() {
// 		// Don't call Unsubscribe here — let Shutdown() handle cleanup
// 		// defer s.bus.Unsubscribe(ch)  // REMOVE THIS LINE
//
// 		for {
// 			select {
// 			case event, ok := <-ch:
// 				if !ok {
// 					log.Println("Email service: event channel closed")
// 					return
// 				}
// 				s.handleEvent(ctx, event)
//
// 			case <-ctx.Done():
// 				log.Println("Email service: shutting down")
// 				return
// 			}
// 		}
// 	}()
// }

func (s *Service) Start(ctx context.Context) {
	ch := s.bus.Subscribe(
		events.EventUserRegistered,
		events.EventPasswordReset,
		events.EventMessageReceived,
	)

	s.wg.Add(1) // register the worker
	go func() {
		defer s.wg.Done()           // signal completion on exit
		defer s.bus.Unsubscribe(ch) // safe now with the closed flag fix

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					log.Println("Email service: event channel closed")
					return
				}
				s.handleEvent(ctx, event)

			case <-ctx.Done():
				log.Println("Email service: shutting down")
				return
			}
		}
	}()
}

// Wait blocks until the email worker goroutine exits
func (s *Service) Wait() {
	s.wg.Wait()
}

func (s *Service) handleEvent(ctx context.Context, event events.Event) {
	switch event.Type {
	case events.EventUserRegistered:
		var payload struct {
			UserID    int64  `json:"user_id"`
			Email     string `json:"email"`
			Name      string `json:"name"`
			VerifyURL string `json:"verify_url"`
		}
		if err := event.PayloadInto(&payload); err != nil {
			log.Printf("Email service: failed to parse user.registered: %v", err)
			return
		}
		s.sendWelcomeEmail(ctx, payload)

	case events.EventPasswordReset:
		var payload struct {
			Email     string `json:"email"`
			ResetURL  string `json:"reset_url"`
			ExpiresAt string `json:"expires_at"`
		}
		if err := event.PayloadInto(&payload); err != nil {
			log.Printf("Email service: failed to parse password_reset: %v", err)
			return
		}
		s.sendPasswordResetEmail(ctx, payload)

	case events.EventMessageReceived:
		var payload struct {
			RecipientEmail string `json:"recipient_email"`
			RecipientName  string `json:"recipient_name"`
			SenderName     string `json:"sender_name"`
			MessagePreview string `json:"message_preview"`
			AppURL         string `json:"app_url"`
		}
		if err := event.PayloadInto(&payload); err != nil {
			log.Printf("Email service: failed to parse message.received: %v", err)
			return
		}
		s.sendNotificationEmail(ctx, payload)
	}
}

func (s *Service) sendWelcomeEmail(ctx context.Context, payload struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	VerifyURL string `json:"verify_url"`
},
) {
	subject := fmt.Sprintf("Welcome to GolangChat, %s!", payload.Name)
	body := fmt.Sprintf(`
Hi %s,

Welcome! Please verify your email by clicking this link:
%s

If you didn't sign up, ignore this email.
`, payload.Name, payload.VerifyURL)

	if err := s.sender.Send(ctx, payload.Email, subject, body, false); err != nil {
		log.Printf("Failed to send welcome email to %s: %v", payload.Email, err)
		// TODO: Retry logic or dead letter queue
	}
}

func (s *Service) sendPasswordResetEmail(ctx context.Context, payload struct {
	Email     string `json:"email"`
	ResetURL  string `json:"reset_url"`
	ExpiresAt string `json:"expires_at"`
},
) {
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
You requested a password reset. Click here to reset:
%s

This link expires at %s.
`, payload.ResetURL, payload.ExpiresAt)

	if err := s.sender.Send(ctx, payload.Email, subject, body, false); err != nil {
		log.Printf("Failed to send password reset to %s: %v", payload.Email, err)
	}
}

func (s *Service) sendNotificationEmail(ctx context.Context, payload struct {
	RecipientEmail string `json:"recipient_email"`
	RecipientName  string `json:"recipient_name"`
	SenderName     string `json:"sender_name"`
	MessagePreview string `json:"message_preview"`
	AppURL         string `json:"app_url"`
},
) {
	subject := fmt.Sprintf("New message from %s", payload.SenderName)
	body := fmt.Sprintf(`
Hi %s,

You have a new message from %s:

"%s"

View it here: %s
`, payload.RecipientName, payload.SenderName, payload.MessagePreview, payload.AppURL)

	if err := s.sender.Send(ctx, payload.RecipientEmail, subject, body, false); err != nil {
		log.Printf("Failed to send notification to %s: %v", payload.RecipientEmail, err)
	}
}
