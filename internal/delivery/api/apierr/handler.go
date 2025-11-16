package apierr

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func Handle(c *gin.Context, err error, opts ...Option) {
	status := http.StatusInternalServerError
	resp := Response{
		Code:    CodeInternalError,
		Message: "Something went wrong!",
	}

	if err != nil {
		if m, ok := DefaultRegistry.FindMapping(err); ok {
			status = m.HTTPStatus
			if m.Message == "" {
				m.Message = err.Error()
			}
			resp = Response{
				Code:    m.Code,
				Message: m.Message,
			}
		}
	}

	for _, o := range opts {
		o(&resp, &status)
	}

	c.AbortWithStatusJSON(status, resp)
}

// Is is a thin wrapper to keep call sites tidy (mirrors errors.Is).
func Is(target, err error) bool { return errors.Is(err, target) }

type Option func(*Response, *int)

func WithMessage(msg string) Option {
	return func(r *Response, _ *int) { r.Message = msg }
}

func WithCode(code string) Option {
	return func(r *Response, _ *int) { r.Code = code }
}

func WithDetail(key string, value any) Option {
	return func(r *Response, _ *int) {
		m, ok := r.Details.(map[string]any)
		if !ok || m == nil {
			m = map[string]any{}
		}
		m[key] = value
		r.Details = m
	}
}

func WithDetails(v any) Option {
	return func(r *Response, _ *int) { r.Details = v }
}

func WithStatus(status int) Option {
	return func(_ *Response, s *int) { *s = status }
}
