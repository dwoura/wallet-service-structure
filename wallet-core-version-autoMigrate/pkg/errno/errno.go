package errno

// Errno defines the error code logic
type Errno struct {
	Code    int
	Message string
}

func (e Errno) Error() string {
	return e.Message
}

// Decode tries to convert an error to Errno
func Decode(err error) (int, string) {
	if err == nil {
		return OK.Code, OK.Message
	}

	switch typed := err.(type) {
	case *Errno:
		return typed.Code, typed.Message
	case Errno:
		return typed.Code, typed.Message
	default:
		return InternalServerError.Code, err.Error()
	}
}

// Common Errors
var (
	OK                  = Errno{Code: 0, Message: "Success"}
	InternalServerError = Errno{Code: 10001, Message: "Internal server error"}
	ErrBind             = Errno{Code: 10002, Message: "Error occurred while binding the request body to the struct"}
	ErrTokenInvalid     = Errno{Code: 10003, Message: "Token invalid"}
	ErrDatabase         = Errno{Code: 10004, Message: "Database error"}
)

// Business Errors (20000+)
var (
	ErrUserNotFound      = Errno{Code: 20101, Message: "User not found"}
	ErrPasswordIncorrect = Errno{Code: 20102, Message: "Password incorrect"}
	ErrAddressNotFound   = Errno{Code: 20201, Message: "Address not found"}
)
