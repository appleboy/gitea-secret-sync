package main

func ToBool(s string) bool {
	if s == "" {
		return false
	}
	return s == "true"
}

func getDataFromEnv(keys []string) map[string]string {
	keysMap := make(map[string]string)
	for _, key := range keys {
		val := getGlobalValue(key)
		if val == "" {
			continue
		}
		keysMap[key] = val
	}
	return keysMap
}
