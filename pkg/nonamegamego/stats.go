package nonamegamego

import (
	"fmt"
	"strings"
)

type Stats struct {
	players []*Player
}

func NewStats(num int) *Stats {
	s := Stats{}

	s.players = make([]*Player, num)
	for i := 0; i < num; i++ {
		s.players[i] = &Player{
			name:   fmt.Sprintf("Игрок-%d", i+1),
			score:  100,
			active: func() bool { return i == 0 }(),
		}
	}

	return &s
}

func (s *Stats) SetActive(num int) {
	for i, p := range s.players {
		p.SetActive(i == num)
	}
}

func (s *Stats) SetActiveNext() (last bool) {
	n := s.ActivePlayerNumber()
	n++
	if n >= len(s.players) {
		n = 0
		last = true
	}
	for i, p := range s.players {
		p.SetActive(i == n)
	}
	return
}

func (s *Stats) ActivePlayer() *Player {
	for _, p := range s.players {
		if p.active {
			return p
		}
	}
	return nil
}

func (s *Stats) ActivePlayerNumber() int {
	for i, p := range s.players {
		if p.active {
			return i
		}
	}
	return 0
}

func (s *Stats) String() string {
	var builder strings.Builder
	builder.WriteString("┌Статистика:──────────┐")
	builder.WriteString("<br>")
	for _, p := range s.players {
		builder.WriteString(fmt.Sprintf("| %s |", p.String()))
		builder.WriteString("<br>")
	}
	builder.WriteString("└─────────────────────┘")
	builder.WriteString("<br>")
	return builder.String()
}
