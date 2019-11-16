package data

import "regexp"

var (
	variableRegex = regexp.MustCompile(`{{[a-zA-Z_][a-zA-Z0-9_\.:-]*}}`)
)
