// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

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
func SimpleStore() Store {
	s := &simpleStore{
		packets: make(chan *packet.PublishPacket),
	}
	n := 2 * runtime.NumCPU()
	for i := 0; i < n; i++ {
		go s.work()
	}
	stores = append(stores, s)
	return s
}

type simpleStore struct {
	sessions sync.Map
	packets  chan *packet.PublishPacket
}

func (s *simpleStore) Count() (count uint64) {
	s.sessions.Range(func(_ interface{}, _ interface{}) bool {
		count++
		return true
	})
	return
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
	s.packets <- pkt
}

func (s *simpleStore) work() {
	for pkt := range s.packets {
		s.sessions.Range(func(_ interface{}, value interface{}) bool {
			session := value.(Session)
			session.Publish(pkt)
			return true
		})
	}
}
