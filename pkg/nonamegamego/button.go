package nonamegamego

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type Button struct {
	Link   string
	Params url.Values
	Text   string
	Class  string
}

type Buttons []Button

func (b *Button) String() string {
	var params string
	if b.Params != nil {
		params = "?" + b.Params.Encode()
	}

	var class string
	if b.Class != "" {
		class = fmt.Sprintf(` class="%s"`, b.Class)
	}

	return fmt.Sprintf(
		`<a href="/spotify/nonamegamego/%s%s"%s>%s</a>`,
		b.Link, params, class, b.Text)
}

func (b *Buttons) Join(sep string) string {
	a := make([]string, len(*b))
	for i, v := range *b {
		a[i] = v.String()
	}
	sort.Strings(a)
	return strings.Join(a, sep)
}
