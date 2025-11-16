package apierr

// Stable, client-facing codes (use in frontend/mobile; avoid renaming).
const (
	CodeInternalError   = "INTERNAL_ERROR"
	CodeBadRequest      = "BAD_REQUEST"
	CodeUnauthorized    = "UNAUTHORIZED"
	CodeForbidden       = "FORBIDDEN"
	CodeNotFound        = "NOT_FOUND"
	CodeConflict        = "CONFLICT"
	CodeNoChanges       = "NO_CHANGES"
	CodeUnprocessable   = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests = "TOO_MANY_REQUESTS"
	CodeValidationError = "VALIDATION_ERROR"
	CodeInvalidToken    = "INVALID_TOKEN"
	CodeExternalService = "EXTERNAL_SERVICE_ERROR"
)
