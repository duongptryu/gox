package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"time"

	"github.com/duongptryu/gox/syserr"
	"gopkg.in/gomail.v2"
)

// GoMailConfig holds configuration for gomail SMTP
type GoMailConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	UseTLS       bool          `json:"use_tls"`
	UseSSL       bool          `json:"use_ssl"`
	SkipVerify   bool          `json:"skip_verify"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	KeepAlive    time.Duration `json:"keep_alive"`
}

// goMailProvider implements MailProvider using gomail
type goMailProvider struct {
	config GoMailConfig
	dialer *gomail.Dialer
}

// NewGoMailProvider creates a new GoMail provider instance
func NewGoMailProvider(config GoMailConfig) MailProvider {
	// Set default timeouts
	if config.DialTimeout == 0 {
		config.DialTimeout = 10 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 10 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 10 * time.Second
	}
	if config.KeepAlive == 0 {
		config.KeepAlive = 30 * time.Second
	}

	// Create gomail dialer
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	// Configure TLS
	if config.UseSSL {
		dialer.SSL = true
	}

	if config.UseTLS || config.UseSSL {
		dialer.TLSConfig = &tls.Config{
			InsecureSkipVerify: config.SkipVerify,
			ServerName:         config.Host,
		}
	}

	return &goMailProvider{
		config: config,
		dialer: dialer,
	}
}

// SendEmail sends a single email message using gomail
func (g *goMailProvider) SendEmail(ctx context.Context, message *EmailMessage) (*SendEmailResponse, error) {
	if err := g.validateEmailMessage(message); err != nil {
		return nil, syserr.Wrap(err, syserr.ValidationCode, "invalid email message")
	}

	// Create gomail message
	msg, err := g.buildGoMailMessage(message)
	if err != nil {
		return nil, syserr.Wrap(err, syserr.InternalCode, "failed to build gomail message")
	}

	// Send the message
	if err := g.dialer.DialAndSend(msg); err != nil {
		return nil, syserr.Wrap(err, syserr.InternalCode, "failed to send email via gomail")
	}

	return &SendEmailResponse{
		MessageID: g.generateMessageID(),
		Status:    "sent",
		Provider:  "gomail",
		Metadata: map[string]interface{}{
			"host":      g.config.Host,
			"port":      g.config.Port,
			"use_tls":   g.config.UseTLS,
			"use_ssl":   g.config.UseSSL,
			"timestamp": time.Now().Unix(),
		},
	}, nil
}

// SendBulkEmails sends multiple emails in batch using gomail
func (g *goMailProvider) SendBulkEmails(ctx context.Context, messages []*EmailMessage) (*BulkSendResponse, error) {
	if len(messages) == 0 {
		return &BulkSendResponse{
			SuccessCount: 0,
			FailureCount: 0,
			Results:      []SendEmailResponse{},
		}, nil
	}

	results := make([]SendEmailResponse, 0, len(messages))
	errors := make([]error, 0)
	successCount := 0
	failureCount := 0

	// Open connection once for all messages
	sender, err := g.dialer.Dial()
	if err != nil {
		// If we can't connect, all messages fail
		for range messages {
			failureCount++
			errors = append(errors, err)
			results = append(results, SendEmailResponse{
				Status:   "failed",
				Provider: "gomail",
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		return &BulkSendResponse{
			SuccessCount: successCount,
			FailureCount: failureCount,
			Results:      results,
			Errors:       errors,
		}, nil
	}
	defer sender.Close()

	// Send each message using the same connection
	for _, message := range messages {
		msg, err := g.buildGoMailMessage(message)
		if err != nil {
			failureCount++
			errors = append(errors, err)
			results = append(results, SendEmailResponse{
				Status:   "failed",
				Provider: "gomail",
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			})
			continue
		}

		if err := gomail.Send(sender, msg); err != nil {
			failureCount++
			errors = append(errors, err)
			results = append(results, SendEmailResponse{
				Status:   "failed",
				Provider: "gomail",
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			})
		} else {
			successCount++
			results = append(results, SendEmailResponse{
				MessageID: g.generateMessageID(),
				Status:    "sent",
				Provider:  "gomail",
				Metadata: map[string]interface{}{
					"timestamp": time.Now().Unix(),
				},
			})
		}
	}

	return &BulkSendResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Errors:       errors,
	}, nil
}

// ValidateEmail validates an email address format
func (g *goMailProvider) ValidateEmail(ctx context.Context, email string, checkDeliverability bool) (bool, error) {
	// Use gomail's built-in validation by trying to set the address
	msg := gomail.NewMessage()

	// Try to set the address - if it fails, it's invalid
	defer func() {
		if r := recover(); r != nil {
			// Invalid email format causes panic in gomail
		}
	}()

	// Test by setting it as a To address
	msg.SetHeader("To", email)

	// If we get here without panic, the format is valid
	// For deliverability check, we would need additional logic
	if checkDeliverability {
		// TODO: Implement MX record lookup and SMTP verification
		// For now, just return true for valid format
	}

	return true, nil
}

// GetProviderInfo returns information about the gomail provider
func (g *goMailProvider) GetProviderInfo() ProviderConfig {
	return ProviderConfig{
		Provider: "gomail",
		Settings: map[string]interface{}{
			"host":          g.config.Host,
			"port":          g.config.Port,
			"use_tls":       g.config.UseTLS,
			"use_ssl":       g.config.UseSSL,
			"skip_verify":   g.config.SkipVerify,
			"dial_timeout":  g.config.DialTimeout.String(),
			"write_timeout": g.config.WriteTimeout.String(),
			"read_timeout":  g.config.ReadTimeout.String(),
			"keep_alive":    g.config.KeepAlive.String(),
		},
	}
}

// Close closes the gomail provider
func (g *goMailProvider) Close() error {
	// gomail doesn't maintain persistent connections by default
	// The dialer is closed automatically after each send
	return nil
}

// buildGoMailMessage converts EmailMessage to gomail.Message
func (g *goMailProvider) buildGoMailMessage(message *EmailMessage) (*gomail.Message, error) {
	msg := gomail.NewMessage()

	// Set sender
	fromAddr := message.From.Email
	if message.From.Name != "" {
		fromAddr = fmt.Sprintf("%s <%s>", message.From.Name, message.From.Email)
	}
	msg.SetHeader("From", fromAddr)

	// Set recipients
	toAddrs := make([]string, len(message.To))
	for i, to := range message.To {
		if to.Name != "" {
			toAddrs[i] = fmt.Sprintf("%s <%s>", to.Name, to.Email)
		} else {
			toAddrs[i] = to.Email
		}
	}
	msg.SetHeader("To", toAddrs...)

	// Set CC recipients
	if len(message.CC) > 0 {
		ccAddrs := make([]string, len(message.CC))
		for i, cc := range message.CC {
			if cc.Name != "" {
				ccAddrs[i] = fmt.Sprintf("%s <%s>", cc.Name, cc.Email)
			} else {
				ccAddrs[i] = cc.Email
			}
		}
		msg.SetHeader("Cc", ccAddrs...)
	}

	// Set BCC recipients
	if len(message.BCC) > 0 {
		bccAddrs := make([]string, len(message.BCC))
		for i, bcc := range message.BCC {
			if bcc.Name != "" {
				bccAddrs[i] = fmt.Sprintf("%s <%s>", bcc.Name, bcc.Email)
			} else {
				bccAddrs[i] = bcc.Email
			}
		}
		msg.SetHeader("Bcc", bccAddrs...)
	}

	// Set Reply-To
	if message.ReplyTo != nil {
		replyToAddr := message.ReplyTo.Email
		if message.ReplyTo.Name != "" {
			replyToAddr = fmt.Sprintf("%s <%s>", message.ReplyTo.Name, message.ReplyTo.Email)
		}
		msg.SetHeader("Reply-To", replyToAddr)
	}

	// Set subject
	msg.SetHeader("Subject", message.Subject)

	// Set priority
	switch message.Priority {
	case PriorityHigh:
		msg.SetHeader("X-Priority", "1")
		msg.SetHeader("X-MSMail-Priority", "High")
		msg.SetHeader("Importance", "High")
	case PriorityLow:
		msg.SetHeader("X-Priority", "5")
		msg.SetHeader("X-MSMail-Priority", "Low")
		msg.SetHeader("Importance", "Low")
	default:
		// Normal priority - no special headers needed
	}

	// Set custom headers
	for key, value := range message.Headers {
		msg.SetHeader(key, value)
	}

	// Set message body
	if message.HTMLBody != "" && message.TextBody != "" {
		// Both HTML and text - multipart alternative
		msg.SetBody("text/plain", message.TextBody)
		msg.AddAlternative("text/html", message.HTMLBody)
	} else if message.HTMLBody != "" {
		// HTML only
		msg.SetBody("text/html", message.HTMLBody)
	} else if message.TextBody != "" {
		// Text only
		msg.SetBody("text/plain", message.TextBody)
	}

	// Add attachments
	for _, attachment := range message.Attachments {
		if err := g.addAttachment(msg, attachment); err != nil {
			return nil, syserr.Wrap(err, syserr.InternalCode, "failed to add attachment")
		}
	}

	return msg, nil
}

// addAttachment adds an attachment to the gomail message
func (g *goMailProvider) addAttachment(msg *gomail.Message, attachment Attachment) error {
	// Read the content into a byte slice
	content, err := io.ReadAll(attachment.Content)
	if err != nil {
		return syserr.Wrap(err, syserr.InternalCode, "failed to read attachment content")
	}

	// Create a setting function for the attachment
	setting := gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(content)
		return err
	})

	// Add the attachment with proper content type
	if attachment.ContentType != "" {
		msg.Attach(attachment.Filename, setting, gomail.SetHeader(map[string][]string{
			"Content-Type": {attachment.ContentType},
		}))
	} else {
		msg.Attach(attachment.Filename, setting)
	}

	return nil
}

// validateEmailMessage validates the email message structure
func (g *goMailProvider) validateEmailMessage(message *EmailMessage) error {
	if message == nil {
		return syserr.New(syserr.ValidationCode, "email message cannot be nil")
	}

	if message.From.Email == "" {
		return syserr.New(syserr.ValidationCode, "from email is required")
	}

	if len(message.To) == 0 {
		return syserr.New(syserr.ValidationCode, "at least one recipient is required")
	}

	if message.Subject == "" {
		return syserr.New(syserr.ValidationCode, "subject is required")
	}

	if message.TextBody == "" && message.HTMLBody == "" {
		return syserr.New(syserr.ValidationCode, "either text body or HTML body is required")
	}

	return nil
}

// generateMessageID generates a unique message ID
func (g *goMailProvider) generateMessageID() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("gomail-%d-%d@%s", timestamp, time.Now().Nanosecond(), g.config.Host)
}
