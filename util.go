package main

import (
	"net"
	"os"
	"strings"
)

// Case-insensitive strings.HasPrefix().  Pass prefix as already upper-case.
func hasCasePrefix(s, prefix string) (ok bool) {
	plen := len(prefix)
	return len(s) >= plen && strings.ToUpper(s[0:plen]) == prefix
}

// Used to parse args of MAIL and RCPT commands (e.g. "From:<foo@bar>").
// Verify prefix matches and discard it.  If <> are present, keep only
// what's inside them.  Strip spaces off both ends.
func parseArg(s string, prefix string) (ret string, ok bool) {
	if hasCasePrefix(s, prefix) {
		s = s[len(prefix):]

		if start := strings.Index(s, "<"); start >= 0 {
			if end := strings.LastIndex(s, ">"); end > start {
				s = s[start+1 : end]
			}
		}

		return strings.TrimSpace(s), true
	}

	return "", false
}

// Simplified strings.SplitN() that always returns two strings.
func splitTwo(s, sep string) (one, two string) {
	if part := strings.SplitN(s, sep, 2); len(part) == 2 {
		return part[0], part[1]
	}

	return s, ""
}

// Strings are immutable in Go. Assuming underlying byte array is not,
// make an explicit copy of it and convert to a string.
func copyToString(src []byte) (line string) {
	dest := make([]byte, len(src))
	copy(dest, src)
	line = string(dest)
	return
}

// Convenience function to see if an arbitrary error object is a timeout.
func isTimeout(err error) (timeout bool) {
	if err == nil {
		return false
	}

	opErr, opOk := err.(*net.OpError)
	timeout = opOk && opErr.Timeout()
	return
}

// Return os.Hostname or "localhost" as string.
func hostname() string {
	if host, err := os.Hostname(); err == nil {
		return host
	}
	return "localhost"
}
