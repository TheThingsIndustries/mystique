// Copyright © 2018 The Things Industries, distributed under the MIT license (see LICENSE file)

package inspect

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/TheThingsIndustries/mystique/pkg/session"
)

type sessionsData struct {
	Sessions sortedSessions `json:"sessions"`
}

func (d sessionsData) sort() { sort.Sort(d.Sessions) }

type sortedSessions []sessionData

func (s sortedSessions) Len() int      { return len(s) }
func (s sortedSessions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortedSessions) Less(i, j int) bool {
	switch {

	// First by ServerName
	case s[i].ServerName < s[j].ServerName:
		return true
	case s[i].ServerName > s[j].ServerName:
		return false

	// Then by Username
	case s[i].Username < s[j].Username:
		return true
	case s[i].Username > s[j].Username:
		return false

	// Then by ClientID
	case s[i].ClientID < s[j].ClientID:
		return true
	case s[i].ClientID > s[j].ClientID:
		return false

	// Finally by RemoteAddr
	default:
		return s[i].RemoteAddr < s[j].RemoteAddr

	}
}

type sessionData struct {
	Transport     string          `json:"transport,omitempty"`
	ServerName    string          `json:"server_name,omitempty"`
	ClientID      string          `json:"client_id"`
	Username      string          `json:"username,omitempty"`
	RemoteAddr    string          `json:"remote_addr"`
	Published     uint64          `json:"published"`
	Delivered     uint64          `json:"delivered"`
	Subscriptions map[string]byte `json:"subscriptions"`
}

// Sessions inspector
func Sessions(s session.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data sessionsData
		for _, sess := range s.All() {
			stats := sess.Stats()
			data.Sessions = append(data.Sessions, sessionData{
				Transport:     sess.AuthInfo().Transport,
				ServerName:    sess.AuthInfo().ServerName,
				ClientID:      sess.AuthInfo().ClientID,
				Username:      sess.AuthInfo().Username,
				RemoteAddr:    sess.AuthInfo().RemoteAddr,
				Published:     stats.Published,
				Delivered:     stats.Delivered,
				Subscriptions: sess.Subscriptions(),
			})
		}
		data.sort()
		if data.Sessions == nil {
			data.Sessions = make(sortedSessions, 0)
		}
		out, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		} else {
			w.Header().Set("content-type", "application/json")
			w.Write(out)
		}
	})
}
