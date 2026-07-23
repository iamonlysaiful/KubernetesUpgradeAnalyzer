package app

type ErrorCategory string

const (
	ErrorUsage         ErrorCategory = "usage"
	ErrorUnimplemented ErrorCategory = "unimplemented"
	ErrorExecution     ErrorCategory = "execution"
)

type AppError struct {
	Category ErrorCategory
	Message  string
	Code     int
	Cause    error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func UsageError(message string) *AppError {
	return &AppError{
		Category: ErrorUsage,
		Message:  message,
		Code:     ExitUsage,
	}
}

func UnimplementedError(command string) *AppError {
	return &AppError{
		Category: ErrorUnimplemented,
		Message:  binaryName + " " + command + " is not implemented yet",
		Code:     ExitExecution,
	}
}

func ExecutionError(message string, cause error) *AppError {
	return &AppError{
		Category: ErrorExecution,
		Message:  message,
		Code:     ExitExecution,
		Cause:    cause,
	}
}
