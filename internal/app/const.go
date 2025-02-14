package app

var (
	ErrBadRequest          = &LogicError{onlyRead: true, code: 400, message: "bad request"}
	ErrUnauthorized        = &LogicError{onlyRead: true, code: 401, message: "unauthorized"}
	ErrForbidden           = &LogicError{onlyRead: true, code: 403, message: "forbidden"}
	ErrNotFound            = &LogicError{onlyRead: true, code: 404, message: "not found"}
	ErrNotAcceptable       = &LogicError{onlyRead: true, code: 406, message: "not acceptable"}
	ErrConflict            = &LogicError{onlyRead: true, code: 409, message: "conflict"}
	ErrUnprocessableEntity = &LogicError{onlyRead: true, code: 422, message: "unprocessable entity"}
	ErrInternal            = &LogicError{onlyRead: true, code: 500, message: "internal error"}
	ErrServiceUnavailable  = &LogicError{onlyRead: true, code: 503, message: "service unavailable"}

	// Ошибка транзакции
	ErrTxСoncurrentExec = NewErrorFrom(ErrInternal)

	// Стандартный ошибки при работе с хранилищами
	ErrStoreUniqueViolation     = NewErrorFrom(ErrConflict)
	ErrStoreForeignKeyViolation = NewErrorFrom(ErrBadRequest)
	ErrStoreCheckViolation      = NewErrorFrom(ErrBadRequest)
	ErrStoreNotNullViolation    = NewErrorFrom(ErrBadRequest)
	ErrStoreRestrictViolation   = NewErrorFrom(ErrBadRequest)
	ErrStoreIntegrityViolation  = NewErrorFrom(ErrBadRequest)
)
