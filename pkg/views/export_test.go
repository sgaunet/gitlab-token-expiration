package views

import "time"

// SetNow sets the clock used to compute expires_in_days, so that tests do not
// depend on the current date.
func SetNow(s *StructuredOutput, now func() time.Time) {
	s.now = now
}
