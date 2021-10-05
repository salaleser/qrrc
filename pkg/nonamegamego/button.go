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
	return fmt.Sprintf("<a href=/spotify/nonamegamego/%s?%s>%s</a>",
		b.Link, b.Params.Encode(), b.Text)
}

func (b *Buttons) Join(sep string) string {
	var builder strings.Builder
	for _, v := range *b {
		builder.WriteString(v.String())
		builder.WriteString(sep)
	}
	return builder.String()
}
