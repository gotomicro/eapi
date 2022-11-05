package view

var (
	ErrInvalidArgument = NewError("invalid_argument")
)

type Error struct {
	Code string `json:"code"`
}

func NewError(code string) *Error {
	return &Error{Code: code}
}
