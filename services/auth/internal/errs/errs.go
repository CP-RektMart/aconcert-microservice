package errs

import "github.com/cockroachdb/errors"

var (
	ErrNotFound = errors.New("not found")
)
