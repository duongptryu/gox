package syserr

import "errors"

func GetStackFormattedFromGenericError(err error) []string {
	var sysErr *Error
	if errors.As(err, &sysErr) {
		return sysErr.StackFormatted()
	}

	return formatStack(extractStackFromGenericError(err))
}

func GetCodeFromGenericError(err error) Code {
	if err == nil {
		return InternalCode
	}

	for {
		var sErr *Error
		if errors.As(err, &sErr) {
			return sErr.Code()
		}

		var unwrapError interface{ Unwrap() error }
		if errors.As(err, &unwrapError) {
			err = unwrapError.Unwrap()
			if err == nil {
				return InternalCode
			}
			continue
		}
		return InternalCode
	}
}

func GetFieldsFromGenericError(err error) []*Field {
	var result []*Field

	for {
		if err == nil {
			return result
		}

		var sErr *Error
		if errors.As(err, &sErr) {
			result = append(result, sErr.Fields()...)
		}

		var unwrapError interface{ Unwrap() error }
		if errors.As(err, &unwrapError) {
			err = unwrapError.Unwrap()
			if err == nil {
				return result
			}
			continue
		}
		return result
	}
}

func UnwrapError(err error) error {
	if err == nil {
		return nil
	}

	for {
		var sErr *Error
		if errors.As(err, &sErr) {
			err = sErr.Unwrap()
			continue
		}

		return err
	}
}
