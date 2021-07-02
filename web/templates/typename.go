package templates

import "strings"

type TypeName string

func (t TypeName) PackagePath() string {
	s := string(t)
	n := strings.LastIndexByte(s, '/')
	if n == -1 {
		return ""
	}

	return s[:n]
}

func (t TypeName) Name() string {
	s := string(t)
	n := strings.LastIndexByte(s, '/')
	return s[n+1:]
}
