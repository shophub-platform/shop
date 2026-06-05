package redis

import "strings"

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func containsInsensitive(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
