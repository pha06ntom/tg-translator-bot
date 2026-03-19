package files

import "os"

func readTextFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	s := string(b)
	return trimToMax(s), nil
}

func trimToMax(s string) string {
	if len([]rune(s)) > maxExtractedChars {
		return string([]rune(s)[:maxExtractedChars])
	}
	return s
}
