package view

var (
	ErrInvalidArgument = NewError(CodeInvalidArgument, "invalid_argument")
)

type ErrCode int

const (
	// Resource not found
	CodeNotFound ErrCode = iota + 10000
	// Request canceld
	CodeCancled
	CodeUnknown
	CodeInvalidArgument
)

type Error struct {
	Code ErrCode `json:"code"`
	Msg  string  `json:"msg"`
}

func NewError(code ErrCode, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}
