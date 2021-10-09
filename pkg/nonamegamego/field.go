package nonamegamego

import (
	"strings"
)

type Field struct {
	Text string
}

type Fields []Field

func (f *Field) String() string {
	return `<input class="name" type="text">`
}

func (f *Fields) Join(sep string) string {
	var builder strings.Builder
	for _, v := range *f {
		builder.WriteString(v.String())
		builder.WriteString(sep)
	}
	result := builder.String()
	return strings.TrimRight(result, sep)
}
