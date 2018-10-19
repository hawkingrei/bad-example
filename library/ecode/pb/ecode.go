package pb

import (
	"strconv"

	"go-common/library/ecode"

	any "github.com/golang/protobuf/ptypes/any"
)

func (e *Error) Error() string {
	return strconv.FormatInt(int64(e.GetErrCode()), 10)
}

// Code is the code of error.
func (e *Error) Code() int {
	return int(e.GetErrCode())
}

// Message is error message.
func (e *Error) Message() string {
	return e.GetErrMessage()
}

// Equal compare whether two errors are equal.
func (e *Error) Equal(ec error) bool {
	return ecode.Cause(ec).Code() == e.Code()
}

// Detail compare whether two errors are equal.
func (e *Error) Detail() interface{} {
	return e.GetErrDetail()
}

// From will convert ecode.Error to pb.Error.
func From(ec ecode.Error) *Error {
	var detail *any.Any
	if d, ok := ec.Detail().(*any.Any); ok && d != nil {
		detail = d
	}
	return &Error{
		ErrCode:    int32(ec.Code()),
		ErrMessage: ec.Message(),
		ErrDetail:  detail,
	}
}
