package metrics

import (
	"fmt"
	"strings"
)

func validateMetric(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("metric cannot be empty")
	}
	n := strings.IndexByte(s, '{')
	if n < 0 {
		return validateIdent(s)
	}
	ident := s[:n]
	s = s[n+1:]
	if err := validateIdent(ident); err != nil {
		return err
	}
	if len(s) == 0 || s[len(s)-1] != '}' {
		return fmt.Errorf("missing closing curly brace at the end of %q", ident)
	}
	return validateTags(s[:len(s)-1])
}

func validateTags(s string) error {
	if len(s) == 0 {
		return nil
	}
	for {
		n := strings.IndexByte(s, '=')
		if n < 0 {
			return fmt.Errorf("missing `=` after %q", s)
		}
		ident := s[:n]
		s = s[n+1:]
		if err := validateIdent(ident); err != nil {
			return err
		}
		if len(s) == 0 || s[0] != '"' {
			return fmt.Errorf("missing starting `\"` for %q value; tail=%q", ident, s)
		}
		s = s[1:]
	again:
		n = strings.IndexByte(s, '"')
		if n < 0 {
			return fmt.Errorf("missing trailing `\"` for %q value; tail=%q", ident, s)
		}
		m := n
		for m > 0 && s[m-1] == '\\' {
			m--
		}
		if (n-m)%2 == 1 {
			s = s[n+1:]
			goto again
		}
		s = s[n+1:]
		if len(s) == 0 {
			return nil
		}
		if !strings.HasPrefix(s, ",") {
			return fmt.Errorf("missing `,` after %q value; tail=%q", ident, s)
		}
		s = skipSpace(s[1:])
	}
}

func skipSpace(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	return s
}

func validateIdent(s string) error {
	if !validateIdentFast(s) {
		return fmt.Errorf("invalid identifier %q", s)
	}
	return nil
}

func validateIdentFast(s string) bool {
	if len(s) == 0 {
		return false
	}

	if !isAlpha(s[0]) && !isSymbol(s[0]) {
		return false
	}

	for i := 1; i < len(s); i++ {
		if !isAlpha(s[i]) && !isNumeric(s[i]) && !isSymbol(s[i]) {
			return false
		}
	}
	return true
}

func isAlpha(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isNumeric(c byte) bool {
	return '0' <= c && c <= '9'
}

func isSymbol(c byte) bool {
	switch c {
	case '_', ':', '.':
		return true
	}
	return false
}
