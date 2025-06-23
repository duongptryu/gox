package mail

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestGoMailSendEmail tests sending a real email via GoMail SMTP
func TestGoMailSendEmail(t *testing.T) {
	// Skip this test by default to avoid sending emails accidentally
	// Remove this line when you want to actually send a test email

	// GoMail Configuration - modify these settings for your SMTP server
	config := GoMailConfig{
		Host:         "smtp.gmail.com", // Change to your SMTP server
		Port:         587,              // Common ports: 25, 465 (SSL), 587 (TLS)
		Username:     username,         // From constants in smtp_test.go
		Password:     password,         // From constants in smtp_test.go
		UseTLS:       true,             // Set to true for STARTTLS
		UseSSL:       false,            // Set to true for direct SSL connection
		SkipVerify:   false,            // Set to true to skip TLS verification
		DialTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		KeepAlive:    30 * time.Second,
	}

	// Create GoMail provider
	provider := NewGoMailProvider(config)
	defer provider.Close()

	// Test email message - CHANGE THE RECIPIENT EMAIL TO YOUR EMAIL
	message := &EmailMessage{
		From: EmailAddress{
			Email: "duongptryu@gmail.com", // Change to your sender email
			Name:  "GoMail Provider Test",
		},
		To: []EmailAddress{
			{
				Email: "duongpt2503@gmail.com", // CHANGE THIS TO YOUR EMAIL ADDRESS
				Name:  "Test Recipient",
			},
		},
		Subject:  "Test Email from GoMail Provider",
		TextBody: "This is a test email sent from the GoMail provider implementation.\n\nIf you receive this, the GoMail implementation is working correctly!",
		HTMLBody: `
			<html>
			<body>
				<h1>GoMail Test Email</h1>
				<p>This is a test email sent from the <strong>GoMail provider</strong> implementation.</p>
				<p>If you receive this, the GoMail implementation is working correctly!</p>
				<p>Key features tested:</p>
				<ul>
					<li>HTML and Text multipart email</li>
					<li>Proper email formatting</li>
					<li>SMTP with TLS/SSL support</li>
					<li>Error handling and validation</li>
				</ul>
				<hr>
				<p><em>Sent at: ` + time.Now().Format(time.RFC3339) + `</em></p>
			</body>
			</html>
		`,
		Priority: PriorityNormal,
		Headers: map[string]string{
			"X-Test-Header": "GoMail Provider Test",
			"X-Mailer":      "GoMail v2 Provider",
		},
	}

	// Send the email
	ctx := context.Background()
	response, err := provider.SendEmail(ctx, message)

	// Check result
	if err != nil {
		t.Fatalf("Failed to send email via GoMail: %v", err)
	}

	// Verify response
	if response == nil {
		t.Fatal("Response is nil")
	}

	if response.Status != "sent" {
		t.Errorf("Expected status 'sent', got '%s'", response.Status)
	}

	if response.Provider != "gomail" {
		t.Errorf("Expected provider 'gomail', got '%s'", response.Provider)
	}

	if response.MessageID == "" {
		t.Error("MessageID should not be empty")
	}

	t.Logf("Email sent successfully via GoMail!")
	t.Logf("Message ID: %s", response.MessageID)
	t.Logf("Status: %s", response.Status)
	t.Logf("Provider: %s", response.Provider)
	t.Logf("Host: %s", response.Metadata["host"])
	t.Logf("Port: %v", response.Metadata["port"])
}

// TestGoMailSendEmailWithCC tests sending an email with CC and BCC recipients
func TestGoMailSendEmailWithCC(t *testing.T) {
	t.Skip("Skipping GoMail CC/BCC test - remove this line to enable")

	config := GoMailConfig{
		Host:        "smtp.gmail.com",
		Port:        587,
		Username:    username,
		Password:    password,
		UseTLS:      true,
		UseSSL:      false,
		DialTimeout: 10 * time.Second,
	}

	provider := NewGoMailProvider(config)
	defer provider.Close()

	message := &EmailMessage{
		From: EmailAddress{
			Email: "duongptryu@gmail.com",
			Name:  "GoMail CC Test",
		},
		To: []EmailAddress{
			{
				Email: "duongpt2503@gmail.com", // Primary recipient
				Name:  "Primary Recipient",
			},
		},
		CC: []EmailAddress{
			{
				Email: "duongptryu@gmail.com", // CC recipient
				Name:  "CC Recipient",
			},
		},
		BCC: []EmailAddress{
			{
				Email: "duongptryu@gmail.com", // BCC recipient (hidden)
				Name:  "BCC Recipient",
			},
		},
		ReplyTo: &EmailAddress{
			Email: "duongptryu@gmail.com",
			Name:  "Reply To Address",
		},
		Subject:  "GoMail Test with CC/BCC",
		TextBody: "This email tests CC and BCC functionality with GoMail provider.",
		HTMLBody: "<p>This email tests <strong>CC and BCC</strong> functionality with GoMail provider.</p>",
		Priority: PriorityHigh,
	}

	ctx := context.Background()
	response, err := provider.SendEmail(ctx, message)

	if err != nil {
		t.Fatalf("Failed to send email with CC/BCC: %v", err)
	}

	t.Logf("Email with CC/BCC sent successfully! Message ID: %s", response.MessageID)
}

// TestGoMailSendEmailWithAttachment tests sending an email with attachment
func TestGoMailSendEmailWithAttachment(t *testing.T) {
	t.Skip("Skipping GoMail attachment test - remove this line to enable")

	config := GoMailConfig{
		Host:        "smtp.gmail.com",
		Port:        587,
		Username:    username,
		Password:    password,
		UseTLS:      true,
		UseSSL:      false,
		DialTimeout: 10 * time.Second,
	}

	provider := NewGoMailProvider(config)
	defer provider.Close()

	// Create a simple text attachment
	attachmentContent := strings.NewReader("This is a test attachment file created by GoMail provider.\nIt contains sample text content for testing purposes.")

	message := &EmailMessage{
		From: EmailAddress{
			Email: "duongptryu@gmail.com",
			Name:  "GoMail Attachment Test",
		},
		To: []EmailAddress{
			{
				Email: "duongpt2503@gmail.com", // CHANGE THIS TO YOUR EMAIL
				Name:  "Test Recipient",
			},
		},
		Subject:  "GoMail Test Email with Attachment",
		TextBody: "This email contains a test attachment sent via GoMail provider.",
		HTMLBody: "<p>This email contains a <strong>test attachment</strong> sent via GoMail provider.</p>",
		Attachments: []Attachment{
			{
				Filename:    "gomail-test.txt",
				Content:     attachmentContent,
				ContentType: "text/plain",
				Size:        int64(attachmentContent.Len()),
			},
		},
	}

	ctx := context.Background()
	response, err := provider.SendEmail(ctx, message)

	if err != nil {
		t.Fatalf("Failed to send email with attachment via GoMail: %v", err)
	}

	t.Logf("Email with attachment sent successfully via GoMail! Message ID: %s", response.MessageID)
}

// TestGoMailBulkSendEmails tests sending multiple emails using GoMail
func TestGoMailBulkSendEmails(t *testing.T) {
	t.Skip("Skipping GoMail bulk email test - remove this line to enable")

	config := GoMailConfig{
		Host:        "smtp.gmail.com",
		Port:        587,
		Username:    username,
		Password:    password,
		UseTLS:      true,
		UseSSL:      false,
		DialTimeout: 10 * time.Second,
	}

	provider := NewGoMailProvider(config)
	defer provider.Close()

	// Create multiple test messages
	messages := []*EmailMessage{
		{
			From:     EmailAddress{Email: "duongptryu@gmail.com", Name: "GoMail Bulk Test"},
			To:       []EmailAddress{{Email: "duongpt2503@gmail.com", Name: "Recipient 1"}}, // CHANGE THIS
			Subject:  "GoMail Bulk Test Email 1",
			TextBody: "This is the first email in the GoMail bulk test.",
			HTMLBody: "<p>This is the <strong>first email</strong> in the GoMail bulk test.</p>",
		},
		{
			From:     EmailAddress{Email: "duongptryu@gmail.com", Name: "GoMail Bulk Test"},
			To:       []EmailAddress{{Email: "duongpt2503@gmail.com", Name: "Recipient 2"}}, // CHANGE THIS
			Subject:  "GoMail Bulk Test Email 2",
			TextBody: "This is the second email in the GoMail bulk test.",
			HTMLBody: "<p>This is the <strong>second email</strong> in the GoMail bulk test.</p>",
		},
		{
			From:     EmailAddress{Email: "duongptryu@gmail.com", Name: "GoMail Bulk Test"},
			To:       []EmailAddress{{Email: "duongpt2503@gmail.com", Name: "Recipient 3"}}, // CHANGE THIS
			Subject:  "GoMail Bulk Test Email 3",
			TextBody: "This is the third email in the GoMail bulk test.",
			HTMLBody: "<p>This is the <strong>third email</strong> in the GoMail bulk test.</p>",
		},
	}

	ctx := context.Background()
	response, err := provider.SendBulkEmails(ctx, messages)

	if err != nil {
		t.Fatalf("Failed to send bulk emails via GoMail: %v", err)
	}

	if response.SuccessCount != 3 {
		t.Errorf("Expected 3 successful emails, got %d", response.SuccessCount)
	}

	if response.FailureCount != 0 {
		t.Errorf("Expected 0 failed emails, got %d", response.FailureCount)
	}

	t.Logf("GoMail bulk emails sent successfully! Success: %d, Failed: %d", response.SuccessCount, response.FailureCount)
}

// TestGoMailProviderInfo tests getting provider information
func TestGoMailProviderInfo(t *testing.T) {
	config := GoMailConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "test@example.com",
		Password:    "password",
		UseTLS:      true,
		UseSSL:      false,
		SkipVerify:  false,
		DialTimeout: 15 * time.Second,
	}

	provider := NewGoMailProvider(config)
	defer provider.Close()

	info := provider.GetProviderInfo()

	if info.Provider != "gomail" {
		t.Errorf("Expected provider 'gomail', got '%s'", info.Provider)
	}

	expectedHost := config.Host
	if host, ok := info.Settings["host"].(string); !ok || host != expectedHost {
		t.Errorf("Expected host '%s', got '%v'", expectedHost, info.Settings["host"])
	}

	expectedPort := config.Port
	if port, ok := info.Settings["port"].(int); !ok || port != expectedPort {
		t.Errorf("Expected port %d, got %v", expectedPort, info.Settings["port"])
	}

	if useTLS, ok := info.Settings["use_tls"].(bool); !ok || useTLS != config.UseTLS {
		t.Errorf("Expected use_tls %v, got %v", config.UseTLS, info.Settings["use_tls"])
	}

	if useSSL, ok := info.Settings["use_ssl"].(bool); !ok || useSSL != config.UseSSL {
		t.Errorf("Expected use_ssl %v, got %v", config.UseSSL, info.Settings["use_ssl"])
	}

	t.Log("GoMail provider info test passed!")
}

// TestGoMailEmailValidation tests email validation functionality
func TestGoMailEmailValidation(t *testing.T) {
	config := GoMailConfig{
		Host:        "localhost",
		Port:        587,
		Username:    "",
		Password:    "",
		UseTLS:      true,
		UseSSL:      false,
		DialTimeout: 5 * time.Second,
	}

	provider := NewGoMailProvider(config)
	defer provider.Close()

	ctx := context.Background()

	// Test valid emails
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@example.org",
		"123@example.com",
		"duongptryu@gmail.com",
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

	t.Log("GoMail email validation tests passed!")
}

// BenchmarkGoMailSendEmail benchmarks the GoMail send email performance
func BenchmarkGoMailSendEmail(b *testing.B) {
	b.Skip("Skipping GoMail benchmark - remove this line to enable")

	config := GoMailConfig{
		Host:        "localhost",
		Port:        587,
		Username:    "",
		Password:    "",
		UseTLS:      false,
		UseSSL:      false,
		DialTimeout: 5 * time.Second,
	}

	provider := NewGoMailProvider(config)
	defer provider.Close()

	message := &EmailMessage{
		From:     EmailAddress{Email: "bench@example.com", Name: "Benchmark Test"},
		To:       []EmailAddress{{Email: "test@example.com", Name: "Test Recipient"}},
		Subject:  "Benchmark Test Email",
		TextBody: "This is a benchmark test email.",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.SendEmail(ctx, message)
	}
}
