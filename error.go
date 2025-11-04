package health

import "github.com/hashicorp/go-multierror"

// Error represents the result of health checks, containing any errors encountered
// indexed by checker name.
type Error struct {
	Errors  map[string]error
	flatten []error
}

// Error implements the error interface for Status.
func (s *Error) Error() string {
	me := &multierror.Error{Errors: s.flatten}

	return me.Error()
}

// Unwrap returns the list of errors contained in the Status.
func (s *Error) Unwrap() []error {
	return s.flatten
}
