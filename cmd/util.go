package main

func ToBool(s string) bool {
	if s == "" {
		return false
	}
	return s == "true"
}
