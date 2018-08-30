package utils

import "unicode"

// CamelToUnderscore will convert camel style naming to underscore style
func CamelToUnderscore(s string) string {
	r := ""
	var prev rune
	for _, c := range s {
		if unicode.IsLower(prev) && unicode.IsUpper(c) {
			r += "_"
		}
		r += string(unicode.ToLower(c))
		prev = c
	}
	return r
}

// UnderscoreToCamel will convert underscore style naming to camel style
func UnderscoreToCamel(s string) string {
	r := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			i++ // Skip the underscore
			if i < len(s) {
				r += string(unicode.ToUpper(rune(s[i])))
			}
			continue
		}

		r += string(c)
	}
	return r
}
