package util

import "html"

func Esc(s string) string {
	return html.EscapeString(s)
}

func Code(s string) string {
	return "<code>" + Esc(s) + "</code>"
}

func Pre(s string) string {
	return "<pre>" + Esc(s) + "</pre>"
}
