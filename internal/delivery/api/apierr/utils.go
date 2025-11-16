package apierr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Convenience helpers for the most common ad-hoc responses.

func BadRequest(c *gin.Context, msg string, details ...any) {
	opts := []Option{
		WithStatus(http.StatusBadRequest),
		WithCode(CodeBadRequest),
		WithMessage(msg),
	}
	if len(details) > 0 && details[0] != nil {
		opts = append(opts, WithDetails(details[0]))
	}
	Handle(c, nil, opts...)
}

func Unauthorized(c *gin.Context, msg string, details ...any) {
	opts := []Option{
		WithStatus(http.StatusUnauthorized),
		WithCode(CodeUnauthorized),
	}
	if msg != "" {
		opts = append(opts, WithMessage(msg))
	}
	if len(details) > 0 && details[0] != nil {
		opts = append(opts, WithDetails(details[0]))
	}
	Handle(c, nil, opts...)
}

func Forbidden(c *gin.Context, msg string) {
	Handle(c, nil,
		WithStatus(http.StatusForbidden),
		WithCode(CodeForbidden),
		WithMessage(msg),
	)
}

func NotFound(c *gin.Context, msg string, details ...any) {
	opts := []Option{
		WithStatus(http.StatusNotFound),
		WithCode(CodeNotFound),
	}
	if msg != "" {
		opts = append(opts, WithMessage(msg))
	}
	if len(details) > 0 && details[0] != nil {
		opts = append(opts, WithDetails(details[0]))
	}
	Handle(c, nil, opts...)
}

func Conflict(c *gin.Context, msg string, details ...any) {
	opts := []Option{
		WithStatus(http.StatusConflict),
		WithCode(CodeConflict),
		WithMessage(msg),
	}
	if len(details) > 0 && details[0] != nil {
		opts = append(opts, WithDetails(details[0]))
	}
	Handle(c, nil, opts...)
}

func Unprocessable(c *gin.Context, msg string, details ...any) {
	opts := []Option{
		WithStatus(http.StatusUnprocessableEntity),
		WithCode(CodeUnprocessable),
		WithMessage(msg),
	}
	if len(details) > 0 && details[0] != nil {
		opts = append(opts, WithDetails(details[0]))
	}
	Handle(c, nil, opts...)
}

func TooManyRequests(c *gin.Context, msg string) {
	Handle(c, nil,
		WithStatus(http.StatusTooManyRequests),
		WithCode(CodeTooManyRequests),
		WithMessage(msg),
	)
}

func InternalError(c *gin.Context, msg string, details ...any) {
	opts := []Option{
		WithStatus(http.StatusInternalServerError),
		WithCode(CodeInternalError),
		WithMessage(msg),
	}
	if len(details) > 0 && details[0] != nil {
		opts = append(opts, WithDetails(details[0]))
	}
	Handle(c, nil, opts...)
}
