package main

func PtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func PtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
