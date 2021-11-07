package process

import (
	"index/suffixarray"
	"regexp"

	"github.com/fatih/color"
)

func highlight(s string, r *regexp.Regexp) string {
	red := color.New(color.FgRed).SprintFunc()

	index := suffixarray.New([]byte(s))
	res := index.FindAllIndex(r, -1)

	newstr := ""
	old := 0

	for _, v := range res {
		newstr = newstr + s[old:v[0]]
		newstr = newstr + red(s[v[0]:v[1]])
		old = v[1]
	}
	newstr = newstr + s[old:]

	return newstr
}
