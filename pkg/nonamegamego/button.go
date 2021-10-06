package nonamegamego

import (
	"fmt"
	"net/url"
	"strings"
)

type Button struct {
	Link   string
	Params url.Values
	Text   string
}

type Buttons []Button

func (b *Button) String() string {
	var params string
	if b.Params != nil {
		params = "?" + b.Params.Encode()
	}
	return fmt.Sprintf("<a href=/spotify/nonamegamego/%s%s>%s</a>",
		b.Link, params, b.Text)
}

func (b *Buttons) Join(sep string) string {
	var builder strings.Builder
	for _, v := range *b {
		builder.WriteString(v.String())
		builder.WriteString(sep)
	}
	result := builder.String()
	return strings.TrimRight(result, sep)
}
