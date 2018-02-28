package main

import (
	"html"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

var usernameRe = regexp.MustCompile("@\\w+")
var hashtagRe = regexp.MustCompile("#[^\\s]+")
var retweetRe = regexp.MustCompile("RT.?:?")
var urlRe = regexp.MustCompile("https?\\S+")

func normalizeText(text string) string {
	text = html.UnescapeString(text)
	text = usernameRe.ReplaceAllString(text, "")
	text = hashtagRe.ReplaceAllString(text, "")
	text = retweetRe.ReplaceAllString(text, "")
	text = urlRe.ReplaceAllString(text, "")
	text = strings.Replace(text, "\n", " ", -1)
	text = strings.Replace(text, "\r", " ", -1)
	text = strings.Replace(text, "\t", " ", -1)
	text = strings.TrimSpace(text)
	text = norm.NFKC.String(text)
	return text
}
