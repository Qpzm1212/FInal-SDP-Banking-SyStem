package patterns

import "errors"

// Validation error mapped to HTTP 400 in cmd/api/main.go.
var ErrInvalidInput = errors.New("invalid data")
