// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package topic implements MQTT topic matching
package topic

import (
	"errors"
	"strings"
)

// MQTT constants
const (
	Separator      = "/"
	Wildcard       = "#"
	PartWildcard   = "+"
	InternalPrefix = "$"
)

// Split a topic into parts
func Split(topic string) []string {
	return strings.Split(topic, Separator)
}

// Join topic parts
func Join(parts []string) string {
	return strings.Join(parts, Separator)
}

// Match a topic to a filter
func Match(topic, filter string) bool {
	return MatchPath(Split(topic), Split(filter))
}

// MatchPath matches a separated topic to a filter
func MatchPath(topicPath, filterPath []string) bool {
	if strings.HasPrefix(topicPath[0], InternalPrefix) &&
		(filterPath[0] == PartWildcard || filterPath[0] == Wildcard) {
		return false
	}
	for i, part := range topicPath {
		if len(filterPath) <= i {
			return false
		}
		if filterPath[i] == Wildcard {
			return true
		}
		if filterPath[i] != PartWildcard && filterPath[i] != part {
			return false
		}
	}
	return len(filterPath) == len(topicPath)
}

// ValidateTopic validates a topic name
func ValidateTopic(topic string) error {
	if len(topic) == 0 {
		return errors.New("Empty topic")
	}
	if strings.ContainsAny(topic, Wildcard+PartWildcard+"\u0000") {
		return errors.New("Topic contains invalid characters")
	}
	return nil
}

// ValidateFilter validates a topic filter
func ValidateFilter(filter string) error {
	if len(filter) == 0 {
		return errors.New("Empty topic filter")
	}
	if strings.ContainsRune(filter, '\u0000') {
		return errors.New("Topic filter can not contain NUL character")
	}
	parts := Split(filter)
	for i, part := range parts {
		if (strings.ContainsAny(part, Wildcard+PartWildcard) && len(part) != 1) ||
			(part == Wildcard && i != len(parts)-1) {
			return errors.New("Topic filter has wildcard in wrong place")
		}
	}
	return nil
}
