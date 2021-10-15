package nonamegamego

import "fmt"

type Player struct {
	name   string
	score  float64
	prev   float64
	active bool
}

func (p *Player) String() string {
	var s string
	if p.active {
		s = " *"
	}

	return fmt.Sprintf(`%.0f: "<label class="player-name">%s</label>" %+.0f %s`,
		p.score, p.name, p.prev, s)
}

func (p *Player) AddScore(n float64) {
	p.prev = n
	p.score += p.prev
}

func (p *Player) SetActive(active bool) {
	p.active = active
}
