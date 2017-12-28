// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package inspect

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/TheThingsIndustries/mystique/pkg/server"
)

type sessionsData struct {
	Sessions sessionsByID `json:"sessions"`
}

func (d sessionsData) sort() { sort.Sort(d.Sessions) }

type sessionsByID []session

func (s sessionsByID) Len() int           { return len(s) }
func (s sessionsByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sessionsByID) Less(i, j int) bool { return s[i].ID < s[j].ID }

type session struct {
	ID            string          `json:"id"`
	Username      string          `json:"username"`
	RemoteAddr    string          `json:"remote_addr"`
	Published     uint64          `json:"published"`
	Delivered     uint64          `json:"delivered"`
	Subscriptions map[string]byte `json:"subscriptions"`
}

// Sessions inspector
func Sessions(s server.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data sessionsData
		for _, sess := range s.Sessions().All() {
			stats := sess.Stats()
			data.Sessions = append(data.Sessions, session{
				ID:            sess.ID(),
				Username:      sess.Username(),
				RemoteAddr:    sess.RemoteAddr(),
				Published:     stats.Published,
				Delivered:     stats.Delivered,
				Subscriptions: sess.Subscriptions(),
			})
		}
		data.sort()
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
