// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"context"
	"sync"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

// Store interface keeps sessions and handles publishing
type Store interface {
	// Get or create a session
	GetOrCreate(id string) ServerSession

	// Delete a session
	Delete(id string)

	// Publish pkt to all sessions
	Publish(pkt *packet.PublishPacket)
}

// SimpleStore returns a simple (inefficient) Store implementation
func SimpleStore(ctx context.Context) Store {
	return &simpleStore{
		ctx:      ctx,
		sessions: make(map[string]*serverSession),
	}
}

type simpleStore struct {
	mu       sync.RWMutex
	ctx      context.Context
	sessions map[string]*serverSession
}

func (s *simpleStore) GetOrCreate(id string) ServerSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		return sess
	}
	sess := &serverSession{session: newSession(s.ctx)}
	s.sessions[id] = sess
	sessionsGauge.Inc()
	return sess
}

func (s *simpleStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; ok {
		delete(s.sessions, id)
		sessionsGauge.Dec()
	}
}

func (s *simpleStore) Publish(pkt *packet.PublishPacket) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sess := range s.sessions {
		go sess.Publish(pkt)
	}
}
