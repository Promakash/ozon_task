package responses

import "net/http"

type Response interface {
	StatusCode() int
	GetPayload() any
}

type BasicResponse struct {
	Payload    any
	statusCode int
}

func (r BasicResponse) StatusCode() int {
	return r.statusCode
}

func (r BasicResponse) GetPayload() any {
	return r.Payload
}

func OK(payload any) *BasicResponse {
	return &BasicResponse{
		statusCode: http.StatusOK,
		Payload:    payload,
	}
}

type ErrorResponse struct {
	Message    string `json:"message"`
	err        error
	statusCode int
}

func (r ErrorResponse) StatusCode() int {
	return r.statusCode
}

func (r ErrorResponse) GetPayload() any {
	return r
}

func BadRequest(err error) *ErrorResponse {
	return &ErrorResponse{
		statusCode: http.StatusBadRequest,
		Message:    err.Error(),
		err:        err,
	}
}

func NotFound(err error) *ErrorResponse {
	return &ErrorResponse{
		statusCode: http.StatusNotFound,
		Message:    err.Error(),
		err:        err,
	}
}

func MethodNotAllowed(err error) *ErrorResponse {
	return &ErrorResponse{
		statusCode: http.StatusMethodNotAllowed,
		Message:    err.Error(),
		err:        err,
	}
}

const InternalError = "Internal server error"

func Unknown(err error) *ErrorResponse {
	return &ErrorResponse{
		statusCode: http.StatusInternalServerError,
		Message:    InternalError,
		err:        err,
	}
}

const TimeoutError = "Request timeout"

func RequestTimeout(err error) *ErrorResponse {
	return &ErrorResponse{
		statusCode: http.StatusRequestTimeout,
		Message:    TimeoutError,
		err:        err,
	}
}
