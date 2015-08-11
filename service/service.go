package service

import (
	"github.com/monopole/croupier/ifc"
	"sync"
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type impl struct {
	players []*ifc.Player
	balls   []*ifc.Ball
	mu      sync.RWMutex
}

func Make() ifc.GameBuddyServerMethods {
	return &impl{
		players: []*ifc.Player{},
		balls:   []*ifc.Ball{},
	}
}

func (x *impl) Recognize(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.players = append(x.players, &p)
	return nil
}

func (x *impl) Forget(_ *context.T, _ rpc.ServerCall, p ifc.Player) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.removePlayer(&p)
	return nil
}

func findIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func (x *impl) removePlayer(p *ifc.Player) {
	i := findIndex(len(x.players), func(i int) bool { return x.players[i].Id == p.Id })
	if i > -1 {
		x.players = append(x.players[:i], x.players[i+1:]...)
	}
}

func (x *impl) Accept(_ *context.T, _ rpc.ServerCall, b ifc.Ball) error {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.balls = append(x.balls, &b)
	return nil
}
