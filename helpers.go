package email

import (
	"html"
	"regexp"
	"strings"
)

var (
	// HTML tags; assumes correctly formed HTML tags, with properly escaped attribute values
	reHtmlTags = regexp.MustCompile(`<[^>]+>`)
	// whitespace, including \xa0 (ASCII 160; non-breaking space)
	reWhitespace = regexp.MustCompile(`[\s\xa0]+`)
	// \xa0 (ASCII 160) is non-breaking space
	htmlToTextREWhitespace = regexp.MustCompile(`(\s|\xa0|&nbsp;)+`)
	// tags that we want removed completely, including contents
	htmlToTextRETagsRm = regexp.MustCompile(`(?i)<head[^a-z].*</head>|<style[^a-z].*</style>|<script[^a-z].*</script>`)
	// tags that we want convert to line breaks
	htmlToTextRETagsLn = regexp.MustCompile(`(?i)<(/h\d|/p|p|br|/ul|/ol|/li|/div|/table|/td)[^a-z]`)
	// tags that we want convert to space
	htmlToTextRETagsSp = regexp.MustCompile(`(?i)<(/?p|br|/?ul|/?ol|/?li|/?div|/?table|/?td|hr|img)`)
	// the alt text from img tags
	htmlToTextREImgAlt = regexp.MustCompile(`(?is)<img [^>]*alt\s*=\s*"([^"]+)"`)
	// the "href" url from links
	htmlToTextREAHref = regexp.MustCompile(`(?is)<a [^>]*href\s*=\s*"([^"]+)".*</a>`)
)

func htmlToText(src string) string {
	// reduce multiple whitespace chars to single space
	src = htmlToTextREWhitespace.ReplaceAllLiteralString(src, " ")
	// remove these tags completely, including contents
	src = htmlToTextRETagsRm.ReplaceAllString(src, "")
	// make sure we have line breaks before these tags
	src = htmlToTextRETagsLn.ReplaceAllString(src, "\n$0")
	// make sure we have white space before these tags
	src = htmlToTextRETagsSp.ReplaceAllString(src, " $0")
	// extract the alt text from images
	src = htmlToTextREImgAlt.ReplaceAllString(src, "$1$0")
	// extract the "href" url from links
	src = htmlToTextREAHref.ReplaceAllString(src, "$0 [ $1 ] ")
	// strip tags
	src = reHtmlTags.ReplaceAllLiteralString(src, "")
	// convert html entities to UTF-8 characters
	src = html.UnescapeString(src)
	// reduce whitespace again; preserve the number of newline chars, or at least a space
	src = reWhitespace.ReplaceAllStringFunc(src, func(m string) string {
		if n := strings.Count(m, "\n"); n > 0 {
			return strings.Repeat("\n", n)
		}
		return " "
	})
	return strings.TrimSpace(src)
}
