package syserr

type Code string

// System error codes.
const (
	InternalCode        Code = "internal"
	InvalidArgumentCode Code = "invalid_argument"
	NotFoundCode        Code = "not_found"
	ConflictCode        Code = "conflict"
	UnauthorizedCode    Code = "unauthorized"
	ForbiddenCode       Code = "forbidden"
	ValidationCode      Code = "validation_error"
)
