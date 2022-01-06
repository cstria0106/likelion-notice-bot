package scraper

import (
	"fmt"
	"strings"
)

type Notice struct {
	Emoji, Title, Type, Uri string
}

func (n Notice) String() string {
	return fmt.Sprintf("%s %s [%s]", n.Emoji, n.Title, n.Type)
}

func noticeFromArray(s []string, uri string) Notice {
	for i, value := range s {
		s[i] = strings.TrimSpace(value)
	}

	uri = strings.TrimSpace(uri)

	if len(s) == 1 {
		return Notice{
			Title: s[0],
			Uri:   uri,
		}
	}

	if len(s) == 2 {
		return Notice{
			Title: s[0],
			Type:  s[1],
			Uri:   uri,
		}
	}

	if len(s) == 3 {
		return Notice{
			Emoji: s[0],
			Title: s[1],
			Type:  s[2],
			Uri:   uri,
		}
	}

	return Notice{
		Uri: uri,
	}
}

func (n *Notice) GetDetail() {

}
