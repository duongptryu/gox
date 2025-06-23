package mail

import (
	"context"
	"io"
)

// EmailAddress represents an email address with optional name
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string    `json:"filename"`
	Content     io.Reader `json:"-"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size,omitempty"`
}

// EmailMessage represents a complete email message
type EmailMessage struct {
	From        EmailAddress      `json:"from"`
	To          []EmailAddress    `json:"to"`
	CC          []EmailAddress    `json:"cc,omitempty"`
	BCC         []EmailAddress    `json:"bcc,omitempty"`
	ReplyTo     *EmailAddress     `json:"reply_to,omitempty"`
	Subject     string            `json:"subject"`
	TextBody    string            `json:"text_body,omitempty"`
	HTMLBody    string            `json:"html_body,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Priority    Priority          `json:"priority,omitempty"`
}

// Priority represents email priority levels
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
)

// SendEmailResponse represents the response after sending an email
type SendEmailResponse struct {
	MessageID string                 `json:"message_id"`
	Status    string                 `json:"status"`
	Provider  string                 `json:"provider"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BulkSendResponse represents the response after sending bulk emails
type BulkSendResponse struct {
	SuccessCount int                 `json:"success_count"`
	FailureCount int                 `json:"failure_count"`
	Results      []SendEmailResponse `json:"results"`
	Errors       []error             `json:"errors,omitempty"`
}

// ProviderConfig represents configuration for the mail provider
type ProviderConfig struct {
	Provider string                 `json:"provider"`
	Settings map[string]interface{} `json:"settings"`
}

// MailProvider defines the interface for sending emails
type MailProvider interface {
	// SendEmail sends a single email message
	SendEmail(ctx context.Context, message *EmailMessage) (*SendEmailResponse, error)

	// SendBulkEmails sends multiple emails in batch
	SendBulkEmails(ctx context.Context, messages []*EmailMessage) (*BulkSendResponse, error)

	// ValidateEmail validates an email address format and optionally checks deliverability
	ValidateEmail(ctx context.Context, email string, checkDeliverability bool) (bool, error)

	// GetProviderInfo returns information about the mail provider
	GetProviderInfo() ProviderConfig

	// Close closes the mail provider and cleans up resources
	Close() error
}
