// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package session

import (
	"runtime"
	"sync"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
)

// Store interface
type Store interface {
	All() []Session
	Store(Session)
	Delete(Session)
	Publish(pkt *packet.PublishPacket)
}

// SimpleStore returns a simple Store implementation and starts a goroutine that keeps the store clean
func SimpleStore() Store { return &simpleStore{} }

type simpleStore struct {
	sessions sync.Map
}

func (s *simpleStore) All() (sessions []Session) {
	s.sessions.Range(func(_ interface{}, value interface{}) bool {
		sessions = append(sessions, value.(Session))
		return true
	})
	return
}

func (s *simpleStore) Store(session Session) {
	s.sessions.Store(session, session)
}

func (s *simpleStore) Delete(session Session) {
	s.sessions.Delete(session)
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
		session := value.(Session)
		queue <- func() {
			session.Publish(pkt)
		}
		return true
	})
	close(queue)
}
