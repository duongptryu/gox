# syserr Package

This package provides structured error handling utilities for Go applications, enabling rich error information, stack traces, error codes, and metadata fields. It is designed to improve error traceability, debugging, and consistency across your codebase.

## Features (Implemented)

- **Structured Error Type**: Custom `Error` type with message, code, stack trace, fields, and error wrapping.
- **Error Codes**: Type-safe error codes for categorizing errors (e.g., `InternalCode`).
- **Stack Trace Support**: Captures and formats stack traces using `github.com/pkg/errors`.
- **Metadata Fields**: Attach arbitrary key-value fields to errors for additional context.
- **Error Wrapping**: Wrap and unwrap errors while preserving stack and metadata.
- **Helper Functions**: Utilities to extract codes, fields, and stack traces from generic errors.

## Usage Example

```go
import "your/module/pkg/syserr"

// Create a new error with a code and message
err := syserr.New(syserr.InternalCode, "something went wrong", syserr.F("user_id", 123))

// Wrap an existing error
wrapped := syserr.Wrap(err, syserr.InternalCode, "failed to process request")

// Extract code, fields, and stack from any error
genericErr := someFunction()
code := syserr.GetCodeFromGenericError(genericErr)
fields := syserr.GetFieldsFromGenericError(genericErr)
stack := syserr.GetStackFormattedFromGenericError(genericErr)
```

## Possible Future Enhancements

- **Expanded Error Codes**: Add more standard error codes (e.g., NotFound, Validation, Unauthorized, etc.).
- **Error Comparison Utilities**: Functions for comparing and matching error types and codes.
- **Integration with Context**: Attach operation/request IDs or user info from context for better traceability.
- **Custom Error Formatting**: Pluggable formatters for error output (e.g., JSON, log-friendly).
- **Localization Support**: Error messages in multiple languages.
- **Error Aggregation**: Support for aggregating multiple errors.
- **Metrics Integration**: Hooks for error reporting/metrics systems.
- **Improved Documentation**: More usage examples and best practices.

---

Feel free to contribute or suggest additional features!
