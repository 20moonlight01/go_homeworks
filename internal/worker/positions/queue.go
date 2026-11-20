package positions

import (
	"sync"

	"github.com/google/uuid"

	"github.com/20moonlight01/go_homeworks/internal/model"
)

type pendingSet struct {
	mu  sync.Mutex
	set map[uuid.UUID]model.StaleTrip
}

func NewPendingSet() *pendingSet {
	return &pendingSet{set: make(map[uuid.UUID]model.StaleTrip)}
}

func (p *pendingSet) TryAdd(t model.StaleTrip) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.set[t.Id]; ok {
		return false
	}
	p.set[t.Id] = t
	return true
}

func (p *pendingSet) Remove(id uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.set, id)
}

func (p *pendingSet) Contains(id uuid.UUID) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.set[id]
	return ok
}

func (p *pendingSet) Cancel(id uuid.UUID) {
	p.Remove(id)
}
