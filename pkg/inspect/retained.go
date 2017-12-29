// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package inspect

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/retained"
)

type retainedData struct {
	Messages messagesByTopic `json:"messages"`
}

func (d retainedData) sort() { sort.Sort(d.Messages) }

type messagesByTopic []*packet.PublishPacket

func (s messagesByTopic) Len() int           { return len(s) }
func (s messagesByTopic) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s messagesByTopic) Less(i, j int) bool { return s[i].TopicName < s[j].TopicName }

// Retained inspector
func Retained(s retained.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data retainedData
		for _, msg := range s.All() {
			data.Messages = append(data.Messages, msg)
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
