// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package subscription implements MQTT topic subscriptions
package subscription

import (
	"strings"
	"sync"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
)

type subscription struct {
	filter     string
	filterPath []string
	qos        byte
}

func (s subscription) Match(topicPath []string) bool {
	return topic.MatchPath(topicPath, s.filterPath)
}

// List of MQTT topic subscriptions
type List struct {
	mu            sync.RWMutex
	subscriptions []subscription
}

// Add a subscription to the list
func (s *List) Add(filter string, qos byte) (added bool) {
	if len(filter) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sub := subscription{
		filter:     filter,
		filterPath: strings.Split(filter, topic.Separator),
		qos:        qos,
	}
	for i, existing := range s.subscriptions {
		if existing.filter == filter {
			s.subscriptions[i] = sub
			return
		}
	}
	s.subscriptions = append(s.subscriptions, sub)
	subscriptionsGauge.Inc()
	return true
}

// Remove a subscription from the list
func (s *List) Remove(filter string) (removed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.subscriptions {
		if sub.filter == filter {
			s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)
			removed = true
			subscriptionsGauge.Dec()
			return
		}
	}
	return
}

// Clear the subscription list
func (s *List) Clear() {
	s.mu.Lock()
	subscriptionsGauge.Sub(float64(len(s.subscriptions)))
	s.subscriptions = nil
	s.mu.Unlock()
}

// Match the topic to the subscriptions and return the maximum QoS
func (s *List) Match(t string) (qos byte, found bool) {
	if len(t) == 0 {
		return
	}
	topicPath := strings.Split(t, topic.Separator)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscriptions {
		if sub.filter == t || sub.Match(topicPath) {
			found = true
			if sub.qos > qos {
				qos = sub.qos
			}
		}
	}
	return
}

// Matches for the topic
func (s *List) Matches(t string) (matches []string) {
	if len(t) == 0 {
		return
	}
	topicPath := strings.Split(t, topic.Separator)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscriptions {
		if sub.filter == t || sub.Match(topicPath) {
			matches = append(matches, sub.filter)
		}
	}
	return
}

// Topics returns all topics in the subscription list
func (s *List) Topics() (topics []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscriptions {
		topics = append(topics, sub.filter)
	}
	return
}
