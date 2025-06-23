package mail

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/duongptryu/gox/syserr"
)

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string        `json:"host"`
	Port     int           `json:"port"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	UseTLS   bool          `json:"use_tls"`
	UseSSL   bool          `json:"use_ssl"`
	Timeout  time.Duration `json:"timeout"`
}

// smtpProvider implements MailProvider using SMTP
type smtpProvider struct {
	config SMTPConfig
}

// NewSMTPProvider creates a new SMTP mail provider
func NewSMTPProvider(config SMTPConfig) MailProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &smtpProvider{
		config: config,
	}
}

// SendEmail sends a single email message via SMTP
func (s *smtpProvider) SendEmail(ctx context.Context, message *EmailMessage) (*SendEmailResponse, error) {
	if err := s.validateEmailMessage(message); err != nil {
		return nil, syserr.Wrap(err, syserr.ValidationCode, "invalid email message")
	}

	// Build the email content
	emailContent, err := s.buildEmailContent(message)
	if err != nil {
		return nil, syserr.Wrap(err, syserr.InternalCode, "failed to build email content")
	}

	// Get all recipients
	recipients := s.getAllRecipients(message)
	if len(recipients) == 0 {
		return nil, syserr.New(syserr.ValidationCode, "no recipients specified")
	}

	// Send via SMTP
	if err := s.sendViaSMTP(ctx, message.From.Email, recipients, emailContent); err != nil {
		return nil, syserr.Wrap(err, syserr.InternalCode, "failed to send email via SMTP")
	}

	return &SendEmailResponse{
		MessageID: s.generateMessageID(),
		Status:    "sent",
		Provider:  "smtp",
		Metadata: map[string]interface{}{
			"host": s.config.Host,
			"port": s.config.Port,
		},
	}, nil
}

// SendBulkEmails sends multiple emails in batch
func (s *smtpProvider) SendBulkEmails(ctx context.Context, messages []*EmailMessage) (*BulkSendResponse, error) {
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

	for _, message := range messages {
		resp, err := s.SendEmail(ctx, message)
		if err != nil {
			failureCount++
			errors = append(errors, err)
			results = append(results, SendEmailResponse{
				Status:   "failed",
				Provider: "smtp",
			})
		} else {
			successCount++
			results = append(results, *resp)
		}
	}

	return &BulkSendResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Errors:       errors,
	}, nil
}

// ValidateEmail validates an email address format and optionally checks deliverability
func (s *smtpProvider) ValidateEmail(ctx context.Context, email string, checkDeliverability bool) (bool, error) {
	// Basic format validation
	if !s.isValidEmailFormat(email) {
		return false, nil
	}

	// If deliverability check is not requested, return true for valid format
	if !checkDeliverability {
		return true, nil
	}

	// For deliverability check, we would need to implement MX record lookup
	// and potentially SMTP verification, which is complex and not always reliable
	// For now, we'll just return true for valid format
	// TODO: Implement deliverability check if needed
	return true, nil
}

// GetProviderInfo returns information about the SMTP provider
func (s *smtpProvider) GetProviderInfo() ProviderConfig {
	return ProviderConfig{
		Provider: "smtp",
		Settings: map[string]interface{}{
			"host":    s.config.Host,
			"port":    s.config.Port,
			"use_tls": s.config.UseTLS,
			"use_ssl": s.config.UseSSL,
		},
	}
}

// Close closes the SMTP provider and cleans up resources
func (s *smtpProvider) Close() error {
	// SMTP connections are typically short-lived, no persistent connections to close
	return nil
}

// validateEmailMessage validates the email message structure
func (s *smtpProvider) validateEmailMessage(message *EmailMessage) error {
	if message == nil {
		return syserr.New(syserr.ValidationCode, "email message cannot be nil")
	}

	if message.From.Email == "" {
		return syserr.New(syserr.ValidationCode, "from email is required")
	}

	if !s.isValidEmailFormat(message.From.Email) {
		return syserr.New(syserr.ValidationCode, "invalid from email format")
	}

	if len(message.To) == 0 {
		return syserr.New(syserr.ValidationCode, "at least one recipient is required")
	}

	for _, to := range message.To {
		if !s.isValidEmailFormat(to.Email) {
			return syserr.New(syserr.ValidationCode, "invalid recipient email format", syserr.F("email", to.Email))
		}
	}

	if message.Subject == "" {
		return syserr.New(syserr.ValidationCode, "subject is required")
	}

	if message.TextBody == "" && message.HTMLBody == "" {
		return syserr.New(syserr.ValidationCode, "either text body or HTML body is required")
	}

	return nil
}

// isValidEmailFormat validates email format using regex
func (s *smtpProvider) isValidEmailFormat(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// getAllRecipients gets all recipient email addresses (To, CC, BCC)
func (s *smtpProvider) getAllRecipients(message *EmailMessage) []string {
	recipients := make([]string, 0)

	for _, to := range message.To {
		recipients = append(recipients, to.Email)
	}

	for _, cc := range message.CC {
		recipients = append(recipients, cc.Email)
	}

	for _, bcc := range message.BCC {
		recipients = append(recipients, bcc.Email)
	}

	return recipients
}

// buildEmailContent builds the complete email content with headers and body
func (s *smtpProvider) buildEmailContent(message *EmailMessage) (string, error) {
	var content strings.Builder

	// Build headers
	content.WriteString(fmt.Sprintf("From: %s\r\n", s.formatEmailAddress(message.From)))
	content.WriteString(fmt.Sprintf("To: %s\r\n", s.formatEmailAddresses(message.To)))

	if len(message.CC) > 0 {
		content.WriteString(fmt.Sprintf("Cc: %s\r\n", s.formatEmailAddresses(message.CC)))
	}

	if message.ReplyTo != nil {
		content.WriteString(fmt.Sprintf("Reply-To: %s\r\n", s.formatEmailAddress(*message.ReplyTo)))
	}

	content.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	content.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	content.WriteString("MIME-Version: 1.0\r\n")

	// Add custom headers
	for key, value := range message.Headers {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Set priority if specified
	if message.Priority != "" {
		switch message.Priority {
		case PriorityHigh:
			content.WriteString("X-Priority: 1\r\n")
			content.WriteString("Importance: high\r\n")
		case PriorityLow:
			content.WriteString("X-Priority: 5\r\n")
			content.WriteString("Importance: low\r\n")
		}
	}

	// Handle content based on whether we have attachments
	if len(message.Attachments) > 0 {
		boundary := s.generateBoundary()
		content.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

		// Add body parts
		if err := s.addBodyParts(&content, message, boundary); err != nil {
			return "", err
		}

		// Add attachments
		if err := s.addAttachments(&content, message.Attachments, boundary); err != nil {
			return "", err
		}

		content.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Simple content without attachments
		if err := s.addSimpleBody(&content, message); err != nil {
			return "", err
		}
	}

	return content.String(), nil
}

// formatEmailAddress formats an email address with optional name
func (s *smtpProvider) formatEmailAddress(addr EmailAddress) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s>", addr.Name, addr.Email)
	}
	return addr.Email
}

// formatEmailAddresses formats multiple email addresses
func (s *smtpProvider) formatEmailAddresses(addrs []EmailAddress) string {
	formatted := make([]string, len(addrs))
	for i, addr := range addrs {
		formatted[i] = s.formatEmailAddress(addr)
	}
	return strings.Join(formatted, ", ")
}

// addSimpleBody adds body content for emails without attachments
func (s *smtpProvider) addSimpleBody(content *strings.Builder, message *EmailMessage) error {
	if message.HTMLBody != "" && message.TextBody != "" {
		// Both HTML and text - use multipart/alternative
		boundary := s.generateBoundary()
		content.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

		// Text part
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.TextBody)
		content.WriteString("\r\n\r\n")

		// HTML part
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.HTMLBody)
		content.WriteString("\r\n\r\n")

		content.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if message.HTMLBody != "" {
		// HTML only
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.HTMLBody)
	} else {
		// Text only
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.TextBody)
	}

	return nil
}

// addBodyParts adds body parts for emails with attachments
func (s *smtpProvider) addBodyParts(content *strings.Builder, message *EmailMessage, boundary string) error {
	if message.HTMLBody != "" && message.TextBody != "" {
		// Both HTML and text - create nested multipart/alternative
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		altBoundary := s.generateBoundary()
		content.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

		// Text part
		content.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.TextBody)
		content.WriteString("\r\n\r\n")

		// HTML part
		content.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.HTMLBody)
		content.WriteString("\r\n\r\n")

		content.WriteString(fmt.Sprintf("--%s--\r\n\r\n", altBoundary))
	} else if message.HTMLBody != "" {
		// HTML only
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.HTMLBody)
		content.WriteString("\r\n\r\n")
	} else {
		// Text only
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		content.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		content.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		content.WriteString(message.TextBody)
		content.WriteString("\r\n\r\n")
	}

	return nil
}

// addAttachments adds attachment parts to the email
func (s *smtpProvider) addAttachments(content *strings.Builder, attachments []Attachment, boundary string) error {
	for _, attachment := range attachments {
		content.WriteString(fmt.Sprintf("--%s\r\n", boundary))

		contentType := attachment.ContentType
		if contentType == "" {
			contentType = mime.TypeByExtension(attachment.Filename)
			if contentType == "" {
				contentType = "application/octet-stream"
			}
		}

		content.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
		content.WriteString("Content-Transfer-Encoding: base64\r\n")
		content.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", attachment.Filename))

		// Read and encode attachment content
		attachmentData, err := io.ReadAll(attachment.Content)
		if err != nil {
			return syserr.Wrap(err, syserr.InternalCode, "failed to read attachment content", syserr.F("filename", attachment.Filename))
		}

		// Base64 encode the attachment
		encoded := s.base64Encode(attachmentData)
		content.WriteString(encoded)
		content.WriteString("\r\n\r\n")
	}

	return nil
}

// sendViaSMTP sends the email via SMTP
func (s *smtpProvider) sendViaSMTP(ctx context.Context, from string, recipients []string, content string) error {
	// Connect to SMTP server
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var client *smtp.Client
	var err error

	if s.config.UseSSL {
		// Direct SSL connection
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}

		conn, err := tls.Dial("tcp", address, tlsConfig)
		if err != nil {
			return syserr.Wrap(err, syserr.InternalCode, "failed to connect to SMTP server with SSL")
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, s.config.Host)
		if err != nil {
			return syserr.Wrap(err, syserr.InternalCode, "failed to create SMTP client")
		}
	} else {
		// Plain connection, possibly with STARTTLS
		client, err = smtp.Dial(address)
		if err != nil {
			return syserr.Wrap(err, syserr.InternalCode, "failed to connect to SMTP server")
		}

		if s.config.UseTLS {
			tlsConfig := &tls.Config{
				ServerName: s.config.Host,
			}

			if err = client.StartTLS(tlsConfig); err != nil {
				client.Close()
				return syserr.Wrap(err, syserr.InternalCode, "failed to start TLS")
			}
		}
	}

	defer client.Close()

	// Authenticate if credentials are provided
	if s.config.Username != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err = client.Auth(auth); err != nil {
			return syserr.Wrap(err, syserr.UnauthorizedCode, "SMTP authentication failed")
		}
	}

	// Set sender
	if err = client.Mail(from); err != nil {
		return syserr.Wrap(err, syserr.InternalCode, "failed to set sender")
	}

	// Set recipients
	for _, recipient := range recipients {
		if err = client.Rcpt(recipient); err != nil {
			return syserr.Wrap(err, syserr.InternalCode, "failed to set recipient", syserr.F("recipient", recipient))
		}
	}

	// Send data
	writer, err := client.Data()
	if err != nil {
		return syserr.Wrap(err, syserr.InternalCode, "failed to initialize data transfer")
	}

	_, err = writer.Write([]byte(content))
	if err != nil {
		writer.Close()
		return syserr.Wrap(err, syserr.InternalCode, "failed to write email content")
	}

	err = writer.Close()
	if err != nil {
		return syserr.Wrap(err, syserr.InternalCode, "failed to finalize email sending")
	}

	return nil
}

// generateMessageID generates a unique message ID
func (s *smtpProvider) generateMessageID() string {
	return fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), s.config.Host)
}

// generateBoundary generates a unique boundary for multipart content
func (s *smtpProvider) generateBoundary() string {
	return fmt.Sprintf("boundary_%d", time.Now().UnixNano())
}

// base64Encode encodes data to base64 with line breaks
func (s *smtpProvider) base64Encode(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)

	// Add line breaks every 76 characters for proper MIME formatting
	const lineLength = 76
	var result strings.Builder
	for i := 0; i < len(encoded); i += lineLength {
		end := i + lineLength
		if end > len(encoded) {
			end = len(encoded)
		}
		result.WriteString(encoded[i:end])
		if end < len(encoded) {
			result.WriteString("\r\n")
		}
	}

	return result.String()
}
