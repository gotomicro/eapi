package tag

import (
	"regexp"
	"strings"
)

var keyValuePattern = regexp.MustCompile("^(\\w+):\"((\\\\.|[^\"])*)\"")
var whiteSpacePattern = regexp.MustCompile("^\\s+")

func Parse(tag string) (res map[string]string) {
	tag = strings.Trim(tag, "`")

	res = make(map[string]string)
	cursor := 0
	for cursor < len(tag) {
		matched := keyValuePattern.FindStringSubmatch(tag[cursor:])
		if matched != nil {
			cursor += len(matched[0])
			res[matched[1]] = matched[2]
			continue
		}

		whitespace := whiteSpacePattern.FindString(tag[cursor:])
		if whitespace != "" {
			cursor += len(whitespace)
			continue
		}

		// unexpected tag
		break
	}

	return
}
