# Mail Package

The mail package provides a flexible and robust email sending solution with multiple provider implementations. It supports sending single emails, bulk emails, email validation, and attachments.

## Features

- **Multiple Providers**: SMTP and GoMail providers
- **Email Validation**: Format validation with optional deliverability checks
- **Bulk Email Support**: Send multiple emails efficiently
- **Attachments**: Support for file attachments with various content types
- **Rich Email Support**: HTML and text body support with multipart messages
- **Priority Settings**: High, normal, and low priority emails
- **Custom Headers**: Add custom email headers
- **CC/BCC Support**: Carbon copy and blind carbon copy recipients
- **Reply-To Support**: Specify reply-to addresses
- **TLS/SSL Support**: Secure email transmission
- **Error Handling**: Comprehensive error handling with detailed messages

## Available Providers

### 1. SMTP Provider (smtp.go)
A custom SMTP implementation with manual message building.

**Features:**
- Manual SMTP protocol implementation
- Custom message formatting
- Basic TLS/SSL support
- Attachment support via base64 encoding

### 2. GoMail Provider (gomail.go) - **Recommended**
A robust implementation using the `gopkg.in/gomail.v2` library.

**Features:**
- Production-ready gomail library
- Better error handling and validation
- Efficient connection reuse for bulk sending
- Superior attachment handling
- Better multipart message support
- More robust TLS/SSL configuration

## Installation

Add the gomail dependency:
```bash
go get gopkg.in/gomail.v2
```

## Quick Start

### Using GoMail Provider (Recommended)

```go
package main

import (
    "context"
    "time"
    
    "github.com/duongptryu/gox/notification/mail"
)

func main() {
    // Configure GoMail provider
    config := mail.GoMailConfig{
        Host:         "smtp.gmail.com",
        Port:         587,
        Username:     "your-email@gmail.com",
        Password:     "your-app-password",
        UseTLS:       true,
        UseSSL:       false,
        SkipVerify:   false,
        DialTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        ReadTimeout:  10 * time.Second,
        KeepAlive:    30 * time.Second,
    }

    // Create provider
    provider := mail.NewGoMailProvider(config)
    defer provider.Close()

    // Create email message
    message := &mail.EmailMessage{
        From: mail.EmailAddress{
            Email: "sender@example.com",
            Name:  "Sender Name",
        },
        To: []mail.EmailAddress{
            {
                Email: "recipient@example.com",
                Name:  "Recipient Name",
            },
        },
        Subject:  "Test Email",
        TextBody: "This is a test email.",
        HTMLBody: "<p>This is a <strong>test email</strong>.</p>",
        Priority: mail.PriorityNormal,
    }

    // Send email
    ctx := context.Background()
    response, err := provider.SendEmail(ctx, message)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Email sent! Message ID: %s\n", response.MessageID)
}
```

### Using SMTP Provider

```go
package main

import (
    "context"
    "time"
    
    "github.com/duongptryu/gox/notification/mail"
)

func main() {
    // Configure SMTP provider
    config := mail.SMTPConfig{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        UseTLS:   true,
        UseSSL:   false,
        Timeout:  30 * time.Second,
    }

    // Create provider
    provider := mail.NewSMTPProvider(config)
    defer provider.Close()

    // Use same EmailMessage structure as above
    // ... (message creation code same as GoMail example)

    // Send email
    response, err := provider.SendEmail(ctx, message)
    // ... (error handling same as above)
}
```

## Advanced Usage

### Sending Email with Attachments

```go
import (
    "os"
    "strings"
)

// Create attachment from file
file, err := os.Open("document.pdf")
if err != nil {
    panic(err)
}
defer file.Close()

// Or create attachment from string
attachmentContent := strings.NewReader("This is attachment content")

message := &mail.EmailMessage{
    From:     mail.EmailAddress{Email: "sender@example.com", Name: "Sender"},
    To:       []mail.EmailAddress{{Email: "recipient@example.com", Name: "Recipient"}},
    Subject:  "Email with Attachment",
    TextBody: "Please find the attached document.",
    Attachments: []mail.Attachment{
        {
            Filename:    "document.pdf",
            Content:     file, // or attachmentContent
            ContentType: "application/pdf", // or "text/plain"
            Size:        1024, // file size in bytes
        },
    },
}
```

### Sending Bulk Emails

```go
messages := []*mail.EmailMessage{
    {
        From:     mail.EmailAddress{Email: "sender@example.com"},
        To:       []mail.EmailAddress{{Email: "user1@example.com"}},
        Subject:  "Bulk Email 1",
        TextBody: "This is the first email.",
    },
    {
        From:     mail.EmailAddress{Email: "sender@example.com"},
        To:       []mail.EmailAddress{{Email: "user2@example.com"}},
        Subject:  "Bulk Email 2",
        TextBody: "This is the second email.",
    },
}

response, err := provider.SendBulkEmails(ctx, messages)
if err != nil {
    panic(err)
}

fmt.Printf("Sent: %d, Failed: %d\n", response.SuccessCount, response.FailureCount)
```

### Email with CC, BCC, and Reply-To

```go
message := &mail.EmailMessage{
    From: mail.EmailAddress{Email: "sender@example.com", Name: "Sender"},
    To: []mail.EmailAddress{
        {Email: "primary@example.com", Name: "Primary Recipient"},
    },
    CC: []mail.EmailAddress{
        {Email: "cc@example.com", Name: "CC Recipient"},
    },
    BCC: []mail.EmailAddress{
        {Email: "bcc@example.com", Name: "BCC Recipient"},
    },
    ReplyTo: &mail.EmailAddress{
        Email: "noreply@example.com",
        Name:  "No Reply",
    },
    Subject:  "Email with Recipients",
    TextBody: "This email has multiple recipient types.",
    Priority: mail.PriorityHigh,
    Headers: map[string]string{
        "X-Campaign-ID": "newsletter-2024",
        "X-Mailer":      "Custom Mailer v1.0",
    },
}
```

### Email Validation

```go
// Basic format validation
valid, err := provider.ValidateEmail(ctx, "test@example.com", false)
if err != nil {
    panic(err)
}

// With deliverability check (if supported)
valid, err := provider.ValidateEmail(ctx, "test@example.com", true)
```

## Configuration Options

### GoMail Configuration

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Host` | string | SMTP server hostname | Required |
| `Port` | int | SMTP server port | Required |
| `Username` | string | SMTP username | Required |
| `Password` | string | SMTP password | Required |
| `UseTLS` | bool | Use STARTTLS | false |
| `UseSSL` | bool | Use direct SSL connection | false |
| `SkipVerify` | bool | Skip TLS certificate verification | false |
| `DialTimeout` | time.Duration | Connection timeout | 10s |
| `WriteTimeout` | time.Duration | Write timeout | 10s |
| `ReadTimeout` | time.Duration | Read timeout | 10s |
| `KeepAlive` | time.Duration | Keep-alive duration | 30s |

### SMTP Configuration

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Host` | string | SMTP server hostname | Required |
| `Port` | int | SMTP server port | Required |
| `Username` | string | SMTP username | Required |
| `Password` | string | SMTP password | Required |
| `UseTLS` | bool | Use STARTTLS | false |
| `UseSSL` | bool | Use direct SSL connection | false |
| `Timeout` | time.Duration | Connection timeout | 30s |

## Email Message Structure

```go
type EmailMessage struct {
    From        EmailAddress      // Sender (required)
    To          []EmailAddress    // Primary recipients (required)
    CC          []EmailAddress    // Carbon copy recipients (optional)
    BCC         []EmailAddress    // Blind carbon copy recipients (optional)
    ReplyTo     *EmailAddress     // Reply-to address (optional)
    Subject     string            // Email subject (required)
    TextBody    string            // Plain text body (optional)
    HTMLBody    string            // HTML body (optional)
    Attachments []Attachment      // File attachments (optional)
    Headers     map[string]string // Custom headers (optional)
    Priority    Priority          // Email priority (optional)
}
```

**Note**: Either `TextBody` or `HTMLBody` (or both) must be provided.

## Priority Levels

- `PriorityHigh`: High priority email
- `PriorityNormal`: Normal priority email (default)
- `PriorityLow`: Low priority email

## Common SMTP Servers

### Gmail
```go
config := mail.GoMailConfig{
    Host:     "smtp.gmail.com",
    Port:     587,
    UseTLS:   true,
    UseSSL:   false,
}
```

### Outlook/Hotmail
```go
config := mail.GoMailConfig{
    Host:     "smtp-mail.outlook.com",
    Port:     587,
    UseTLS:   true,
    UseSSL:   false,
}
```

### Yahoo
```go
config := mail.GoMailConfig{
    Host:     "smtp.mail.yahoo.com",
    Port:     587,
    UseTLS:   true,
    UseSSL:   false,
}
```

### Custom SMTP
```go
config := mail.GoMailConfig{
    Host:     "smtp.yourdomain.com",
    Port:     25,   // or 465 for SSL, 587 for TLS
    UseTLS:   true,
    UseSSL:   false,
}
```

## Testing

Run tests with:
```bash
# Run all tests (skipped by default)
go test ./notification/mail/

# Run specific provider tests
go test ./notification/mail/ -run TestGoMail
go test ./notification/mail/ -run TestSendEmail

# Run benchmarks
go test ./notification/mail/ -bench=.
```

**Note**: Most tests are skipped by default to avoid sending real emails. Remove the `t.Skip()` lines in test files to enable them, and update the email addresses in the test configurations.

## Error Handling

The mail providers use the `syserr` package for structured error handling:

```go
response, err := provider.SendEmail(ctx, message)
if err != nil {
    // Check error type
    if sysErr, ok := err.(*syserr.Error); ok {
        switch sysErr.Code {
        case syserr.ValidationCode:
            fmt.Println("Validation error:", sysErr.Message)
        case syserr.InternalCode:
            fmt.Println("Internal error:", sysErr.Message)
        default:
            fmt.Println("Unknown error:", sysErr.Message)
        }
    }
    return
}
```

## Provider Comparison

| Feature | SMTP Provider | GoMail Provider |
|---------|---------------|-----------------|
| **Reliability** | Basic | Production-ready |
| **Performance** | Good | Better |
| **Bulk Sending** | Sequential | Connection reuse |
| **Attachment Handling** | Manual base64 | Native support |
| **Multipart Messages** | Manual | Automatic |
| **Error Handling** | Basic | Comprehensive |
| **TLS Configuration** | Basic | Advanced |
| **Message Validation** | Manual | Built-in |
| **Recommended Use** | Simple cases | Production use |

## Best Practices

1. **Use GoMail Provider**: Recommended for production applications
2. **Connection Reuse**: Use bulk sending for multiple emails
3. **Error Handling**: Always check and handle errors appropriately
4. **Timeouts**: Set appropriate timeouts for your use case
5. **TLS/SSL**: Always use TLS or SSL for secure transmission
6. **Email Validation**: Validate email addresses before sending
7. **Rate Limiting**: Implement rate limiting for bulk sending
8. **App Passwords**: Use app-specific passwords for Gmail, Outlook, etc.

## Security Considerations

- Never hardcode credentials in source code
- Use environment variables or secure configuration management
- Enable TLS/SSL for encrypted transmission
- Use app-specific passwords instead of account passwords
- Implement proper access controls for email sending functionality
- Validate all input data to prevent injection attacks

## License

This package is part of the gox framework. See the main project license for details. 