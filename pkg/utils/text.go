package utils

func Plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
