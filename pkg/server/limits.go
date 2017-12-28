// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package server

import (
	"errors"
	"sync"
)

type limits struct {
	sync.Mutex

	max int
	cnt map[string]int
}

func newLimits(max int) *limits {
	return &limits{
		max: max,
		cnt: make(map[string]int),
	}
}

func (l *limits) connect(id string) error {
	if l == nil {
		return nil
	}
	l.Lock()
	defer l.Unlock()
	if l.max > 0 && l.cnt[id] >= l.max {
		return errors.New("limit reached")
	}
	l.cnt[id] = l.cnt[id] + 1
	return nil
}

func (l *limits) disconnect(id string) {
	if l == nil {
		return
	}
	l.Lock()
	defer l.Unlock()
	l.cnt[id] = l.cnt[id] - 1
	if l.cnt[id] == 0 {
		delete(l.cnt, id)
	}
}
