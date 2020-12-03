package main

import "regexp"

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`Код ([\d\w]+)(?:\.|\s|$)`),
	regexp.MustCompile(`Kod: ([\d\w]+)(?:\.|\s|$)`),
	regexp.MustCompile(`(?i)код: ([\d\w]+)(?:\.|\s|$)`),
	regexp.MustCompile(`Код активации Apple Pay ([\d\w]+)`),
	regexp.MustCompile(`Код подтверждения: ([\d\w]+)`),
	regexp.MustCompile(`Код`),
}

func isCode(s string) bool {
	for _, p := range patterns {
		if p.MatchString(s) {
			return true
		}
	}
	return false
}
