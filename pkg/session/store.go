// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

// Store interface keeps sessions and handles publishing
type Store interface {
	All() []ServerSession

	Cleanup()

	// Get or create a session
	GetOrCreate(id string) ServerSession

	// Delete a session
	Delete(id string)

	// Publish pkt to all sessions
	Publish(pkt *packet.PublishPacket)
}

// SimpleStore returns a simple (inefficient) Store implementation and starts a goroutine that keeps the store clean
func SimpleStore(ctx context.Context) Store {
	s := &simpleStore{ctx: ctx}
	stores = append(stores, s)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
				s.Cleanup()
			}
		}
	}()
	return s
}

type simpleStore struct {
	mu       sync.Mutex // GetOrCreate and Cleanup are mutually exclusive
	ctx      context.Context
	sessions sync.Map
}

func (s *simpleStore) All() (sessions []ServerSession) {
	s.sessions.Range(func(_ interface{}, value interface{}) bool {
		if session, ok := value.(*serverSession); ok {
			if !session.IsGarbage() {
				sessions = append(sessions, session)
			}
		}
		return true
	})
	return
}

func (s *simpleStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions.Range(func(idI interface{}, sessionI interface{}) bool {
		id := idI.(string)
		session := sessionI.(*serverSession)
		session.mu.RLock()
		if session.IsGarbage() {
			s.sessions.Delete(id)
		}
		session.mu.RUnlock()
		return true
	})
}

func (s *simpleStore) GetOrCreate(id string) ServerSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionI, _ := s.sessions.LoadOrStore(id, &serverSession{session: newSession(s.ctx)})
	return sessionI.(*serverSession)
}

func (s *simpleStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions.Delete(id)
}

func (s *simpleStore) Publish(pkt *packet.PublishPacket) {
	workers := runtime.NumCPU()
	queue := make(chan func(), workers)
	for i := 0; i < workers; i++ {
		go func() {
			for publish := range queue {
				publish()
			}
		}()
	}
	s.sessions.Range(func(_ interface{}, value interface{}) bool {
		if session, ok := value.(*serverSession); ok {
			if !session.IsGarbage() {
				queue <- func() {
					session.Publish(pkt)
				}
			}
		}
		return true
	})
	close(queue)
}
