package mail

import (
	"context"
	"strings"
	"testing"
	"time"
)

const (
	username = ""
	password = ""
)

// TestSendEmail tests sending a real email via SMTP
// Modify the configuration and recipient email as needed
func TestSendEmail(t *testing.T) {
	// Skip this test by default to avoid sending emails accidentally
	// Remove this line when you want to actually send a test email

	// SMTP Configuration - modify these settings for your SMTP server
	config := SMTPConfig{
		Host:     "smtp.gmail.com", // Change to your SMTP server (e.g., "smtp.gmail.com", "localhost")
		Port:     587,              // Common ports: 25, 465 (SSL), 587 (TLS)
		Username: username,         // Empty as requested
		Password: password,         // Empty as requested
		UseTLS:   true,             // Set to true for STARTTLS, false for plain
		UseSSL:   false,            // Set to true for direct SSL connection
		Timeout:  30 * time.Second,
	}

	// Create SMTP provider
	provider := NewSMTPProvider(config)
	defer provider.Close()

	// Test email message - CHANGE THE RECIPIENT EMAIL TO YOUR EMAIL
	message := &EmailMessage{
		From: EmailAddress{
			Email: "duongptryu@gmail.com", // Change to your sender email
			Name:  "Duong ProMax",
		},
		To: []EmailAddress{
			{
				Email: "duongpt2503@gmail.com", // CHANGE THIS TO YOUR EMAIL ADDRESS
				Name:  "Duong Max Pro",
			},
		},
		Subject:  "Test Email from Go SMTP Provider",
		TextBody: "This is a test email sent from the Go SMTP provider implementation.\n\nIf you receive this, the implementation is working correctly!",
		HTMLBody: `
			<html>
			<body>
				<h1>Test Email</h1>
				<p>This is a test email sent from the <strong>Go SMTP provider</strong> implementation.</p>
				<p>If you receive this, the implementation is working correctly!</p>
				<hr>
				<p><em>Sent at: ` + time.Now().Format(time.RFC3339) + `</em></p>
			</body>
			</html>
		`,
		Priority: PriorityNormal,
		Headers: map[string]string{
			"X-Test-Header": "Go SMTP Test",
		},
	}

	// Send the email
	ctx := context.Background()
	response, err := provider.SendEmail(ctx, message)

	// Check result
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	// Verify response
	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.Status != "sent" {
		t.Errorf("Expected status 'sent', got '%s'", response.Status)
	}

	if response.Provider != "smtp" {
		t.Errorf("Expected provider 'smtp', got '%s'", response.Provider)
	}

	if response.MessageID == "" {
		t.Error("MessageID should not be empty")
	}

	t.Logf("Email sent successfully!")
	t.Logf("Message ID: %s", response.MessageID)
	t.Logf("Status: %s", response.Status)
	t.Logf("Provider: %s", response.Provider)
}

// TestSendEmailWithAttachment tests sending an email with attachment
func TestSendEmailWithAttachment(t *testing.T) {
	// Skip this test by default
	t.Skip("Skipping attachment test - remove this line to enable")

	config := SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "",
		Password: "",
		UseTLS:   true,
		UseSSL:   false,
		Timeout:  30 * time.Second,
	}

	provider := NewSMTPProvider(config)
	defer provider.Close()

	// Create a simple text attachment
	attachmentContent := strings.NewReader("This is a test attachment file.\nCreated by Go SMTP test.")

	message := &EmailMessage{
		From: EmailAddress{
			Email: "test@example.com",
			Name:  "Test Sender",
		},
		To: []EmailAddress{
			{
				Email: "your-email@example.com", // CHANGE THIS TO YOUR EMAIL
				Name:  "Test Recipient",
			},
		},
		Subject:  "Test Email with Attachment",
		TextBody: "This email contains a test attachment.",
		HTMLBody: "<p>This email contains a <strong>test attachment</strong>.</p>",
		Attachments: []Attachment{
			{
				Filename:    "test.txt",
				Content:     attachmentContent,
				ContentType: "text/plain",
				Size:        int64(attachmentContent.Len()),
			},
		},
	}

	ctx := context.Background()
	response, err := provider.SendEmail(ctx, message)

	if err != nil {
		t.Fatalf("Failed to send email with attachment: %v", err)
	}

	t.Logf("Email with attachment sent successfully! Message ID: %s", response.MessageID)
}

// TestBulkSendEmails tests sending multiple emails
func TestBulkSendEmails(t *testing.T) {
	// Skip this test by default
	t.Skip("Skipping bulk email test - remove this line to enable")

	config := SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "",
		Password: "",
		UseTLS:   true,
		UseSSL:   false,
		Timeout:  30 * time.Second,
	}

	provider := NewSMTPProvider(config)
	defer provider.Close()

	// Create multiple test messages
	messages := []*EmailMessage{
		{
			From:     EmailAddress{Email: "test@example.com", Name: "Test Sender"},
			To:       []EmailAddress{{Email: "your-email@example.com", Name: "Recipient 1"}}, // CHANGE THIS
			Subject:  "Bulk Test Email 1",
			TextBody: "This is the first email in the bulk test.",
		},
		{
			From:     EmailAddress{Email: "test@example.com", Name: "Test Sender"},
			To:       []EmailAddress{{Email: "your-email@example.com", Name: "Recipient 2"}}, // CHANGE THIS
			Subject:  "Bulk Test Email 2",
			TextBody: "This is the second email in the bulk test.",
		},
	}

	ctx := context.Background()
	response, err := provider.SendBulkEmails(ctx, messages)

	if err != nil {
		t.Fatalf("Failed to send bulk emails: %v", err)
	}

	if response.SuccessCount != 2 {
		t.Errorf("Expected 2 successful emails, got %d", response.SuccessCount)
	}

	if response.FailureCount != 0 {
		t.Errorf("Expected 0 failed emails, got %d", response.FailureCount)
	}

	t.Logf("Bulk emails sent successfully! Success: %d, Failed: %d", response.SuccessCount, response.FailureCount)
}

// TestEmailValidation tests email validation functionality
func TestEmailValidation(t *testing.T) {
	config := SMTPConfig{
		Host:     "localhost",
		Port:     587,
		Username: "",
		Password: "",
		UseTLS:   true,
		UseSSL:   false,
	}

	provider := NewSMTPProvider(config)
	defer provider.Close()

	ctx := context.Background()

	// Test valid emails
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
		"123@example.com",
	}

	for _, email := range validEmails {
		valid, err := provider.ValidateEmail(ctx, email, false)
		if err != nil {
			t.Errorf("Unexpected error validating %s: %v", email, err)
		}
		if !valid {
			t.Errorf("Expected %s to be valid", email)
		}
	}

	// Test invalid emails
	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"user@",
		"user space@example.com",
		"",
	}

	for _, email := range invalidEmails {
		valid, err := provider.ValidateEmail(ctx, email, false)
		if err != nil {
			t.Errorf("Unexpected error validating %s: %v", email, err)
		}
		if valid {
			t.Errorf("Expected %s to be invalid", email)
		}
	}

	t.Log("Email validation tests passed!")
}

// TestProviderInfo tests getting provider information
func TestProviderInfo(t *testing.T) {
	config := SMTPConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "",
		Password: "",
		UseTLS:   true,
		UseSSL:   false,
	}

	provider := NewSMTPProvider(config)
	defer provider.Close()

	info := provider.GetProviderInfo()

	if info.Provider != "smtp" {
		t.Errorf("Expected provider 'smtp', got '%s'", info.Provider)
	}

	expectedHost := config.Host
	if host, ok := info.Settings["host"].(string); !ok || host != expectedHost {
		t.Errorf("Expected host '%s', got '%v'", expectedHost, info.Settings["host"])
	}

	expectedPort := config.Port
	if port, ok := info.Settings["port"].(int); !ok || port != expectedPort {
		t.Errorf("Expected port %d, got %v", expectedPort, info.Settings["port"])
	}

	t.Log("Provider info test passed!")
}
