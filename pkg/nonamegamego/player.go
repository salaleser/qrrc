package nonamegamego

import "fmt"

type Player struct {
	name   string
	score  int
	active bool
}

func (p *Player) String() string {
	var s string
	if p.active {
		s = " (!)"
	}
	return fmt.Sprintf("%q [%d]%s", p.name, p.score, s)
}

func (p *Player) AddScore(n int) {
	p.score += n
}

func (p *Player) SetActive(active bool) {
	p.active = active
}
