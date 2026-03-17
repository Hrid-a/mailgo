package verifier

import "strings"

// SMTPError is a parsed error from an SMTP response
type SMTPError struct {
	Message string
}

func (e *SMTPError) Error() string {
	return e.Message
}

// ParseSMTPError maps a raw SMTP error to a structured SMTPError with a known message
func ParseSMTPError(err error) *SMTPError {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case insContains(msg, "452", "full", "quota", "insufficient"):
		return &SMTPError{ErrFullInbox}
	case insContains(msg, "550", "551", "553", "does not exist", "no such", "invalid user", "unknown user"):
		return &SMTPError{ErrMailboxNotFound}
	case insContains(msg, "521", "421", "not allowed", "blocked", "denied"):
		return &SMTPError{ErrNotAllowed}
	case insContains(msg, "450", "451", "unverified", "temporarily"):
		return &SMTPError{ErrTempUnavailable}
	default:
		return &SMTPError{err.Error()}
	}
}

// insContains checks if s contains any of the given substrings (case-insensitive)
func insContains(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
