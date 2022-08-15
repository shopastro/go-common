package limiter

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// im memory rate limit

type LimiterStore struct {
	mutex sync.RWMutex
	store map[string]*rate.Limiter
}

func NewLimiterStore() *LimiterStore {
	return &LimiterStore{
		mutex: sync.RWMutex{},
		store: make(map[string]*rate.Limiter),
	}
}

// r => hom many token generated per second
// b => bucket size
func (s *LimiterStore) RegisterLimiter(name string, r float64, b int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[name] = rate.NewLimiter(rate.Limit(r), b)
}

func (s *LimiterStore) GetLimiter(name string) *rate.Limiter {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if limiter, ok := s.store[name]; ok {
		return limiter
	}
	return nil
}

func (s *LimiterStore) UnregisterLimiter(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.store, name)
}

func (s *LimiterStore) UnregisterAllLimiter() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store = make(map[string]*rate.Limiter)
}

// wait will block until 1 token is available or ctx is canceled/deadline reached
func (s *LimiterStore) Wait(ctx context.Context, name string) error {
	if limiter := s.GetLimiter(name); limiter != nil {
		return limiter.Wait(ctx)
	}
	// no limiter available, pass
	return nil
}
