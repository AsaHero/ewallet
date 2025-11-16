package apierr

import (
	"errors"
	"net/http"
	"sync"

	"github.com/AsaHero/e-wallet/internal/inerr"
)

type Mapping struct {
	HTTPStatus int
	Code       string
	Message    string
	Details    map[string]any
}

type regItem struct {
	match   func(error) bool
	builder func(error) Mapping
}

type Registry struct {
	mu   sync.RWMutex
	maps []regItem
}

func NewRegistry() *Registry {
	r := &Registry{}

	// Register mappings for your inerr package (extend as needed).
	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrNotFound{}) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusNotFound,
				Code:       CodeNotFound,
			}
		},
	)

	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrConflict{}) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusConflict,
				Code:       CodeConflict,
			}
		},
	)

	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrNoChanges{}) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusBadRequest,
				Code:       CodeNoChanges,
			}
		},
	)

	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrorExpiredToken) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusUnauthorized,
				Code:       CodeInvalidToken,
				Message:    "token expired",
				Details: map[string]any{
					"error": err.Error(),
				},
			}
		},
	)

	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrorWrongAlgo) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusUnauthorized,
				Code:       CodeInvalidToken,
				Message:    "token expired",
				Details: map[string]any{
					"error": err.Error(),
				},
			}
		},
	)

	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrInvalidToken{}) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusUnauthorized,
				Code:       CodeInvalidToken,
			}
		},
	)

	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrValidation{}) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusBadRequest,
				Code:       CodeValidationError,
			}
		},
	)

	r.RegisterMatch(func(err error) bool { return errors.Is(err, inerr.ErrorPermissionDenied) },
		func(err error) Mapping {
			return Mapping{
				HTTPStatus: http.StatusForbidden,
				Code:       CodeForbidden,
			}
		},
	)

	r.RegisterMatch(
		func(err error) bool {
			return errors.Is(err, inerr.ErrHttp{})
		},
		func(err error) Mapping {
			httpErr, ok := err.(*inerr.ErrHttp)
			if !ok {
				return Mapping{
					HTTPStatus: http.StatusInternalServerError,
					Code:       CodeInternalError,
				}
			}

			return Mapping{
				HTTPStatus: httpErr.StatusCode,
				Code:       CodeExternalService,
				Message:    httpErr.Message,
				Details: map[string]any{
					"body": httpErr.Body,
				},
			}
		},
	)

	return r
}

// RegisterMatch registers how to map an error to an API response.
func (r *Registry) RegisterMatch(match func(error) bool, builder func(error) Mapping) {
	r.mu.Lock()
	r.maps = append(r.maps, regItem{match: match, builder: builder})
	r.mu.Unlock()
}

// FindMapping looks up a mapping on THIS registry.
func (r *Registry) FindMapping(err error) (Mapping, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, it := range r.maps {
		if it.match(err) {
			return it.builder(err), true
		}
	}
	return Mapping{}, false
}

// Singleton
var DefaultRegistry = NewRegistry()
