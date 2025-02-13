package app

import "fmt"

type LogicError struct {
	onlyRead bool
	code     int
	message  string
	details  []string
	wrapErr  []error
	data     any
}

func (e *LogicError) Error() string {
	if len(e.details) > 0 {
		return fmt.Sprintf("%s - details: %v", e.message, e.details)
	}
	return e.message
}

func IsAppError(err error) bool {
	_, ok := err.(*LogicError)
	return ok
}

func NewError(code int, message string, details []string) *LogicError {
	return &LogicError{
		code:    code,
		message: message,
		details: details,
	}
}

func NewErrorFrom(parentErr *LogicError) *LogicError {
	e := &LogicError{
		code:    parentErr.code,
		message: parentErr.message,
		details: parentErr.details,
	}
	return e.Wrap(parentErr)
}

func (e *LogicError) Wrap(err error) *LogicError {
	if e.onlyRead {
		return e
	}

	e.wrapErr = append(e.wrapErr, err)
	return e
}

func (e *LogicError) Unwrap() []error {
	return e.wrapErr
}

func (e *LogicError) Is(err error) bool {
	for _, wErr := range e.wrapErr {
		if wErr == err {
			return true
		}
	}
	return false
}

func (e *LogicError) Code() int {
	return e.code
}

func (e *LogicError) Message() string {
	return e.message
}

func (e *LogicError) Details() []string {
	return e.details
}

func (e *LogicError) SetMessage(message string) *LogicError {
	if e.onlyRead {
		return e
	}
	e.message = message
	return e
}

func (e *LogicError) AddDetails(details []string) *LogicError {
	if e.onlyRead {
		return e
	}
	if e.details == nil {
		return e
	}
	if e.details == nil {
		e.details = make([]string, len(details))
	}
	e.details = append(e.details, details...)
	return e
}

func (e *LogicError) Data() any {
	return e.data
}

func (e *LogicError) SetData(data any) *LogicError {
	if e.onlyRead {
		return e
	}

	e.data = data
	return e
}
