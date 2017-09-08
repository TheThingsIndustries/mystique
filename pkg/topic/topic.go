// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

// Package topic implements MQTT topic matching
package topic

import "strings"

// MQTT constants
const (
	Separator      = "/"
	Wildcard       = "#"
	PartWildcard   = "+"
	InternalPrefix = "$"
)

// Match a topic to a filter
func Match(topic, filter string) bool {
	return MatchPath(strings.Split(topic, Separator), strings.Split(filter, Separator))
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
