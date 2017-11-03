// Copyright © 2017 The Things Industries, distributed under the MIT license (see LICENSE file)

package inspect

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/TheThingsIndustries/mystique/pkg/server"
	"github.com/TheThingsIndustries/mystique/pkg/session"
)

var tmpl = template.Must(template.New("inspect").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Mystique - Inspect Sessions</title>
	<style>
	body {
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
		color: #222;
		font-size: 12pt;
	}
	h1, h2, h3 {
		vertical-align: middle;
		font-family: "League Spartan";
		text-transform: uppercase;
		margin-top: 0.2em;
	}
	code {
		font-family: Consolas, Monaco, "Andale Mono", "Ubuntu Mono", monospace;
		text-transform: none;
	}
	h1 small, h2 small {
		color: #666;
		font-size: 60%;
	}
	h2 small span {
		margin-right: 5px;
		white-space: nowrap;
	}
	section.session {
		border: 1px solid #CCC;
		border-radius: 3px;
		background-color: #FAFAFA;
		padding: 5px;
		margin-bottom: 5px;
	}
	</style>
</head>
<body>
{{ range .Sessions }}
<section class="session">
<h2>Session <code>{{ .ID }}</code><br><small>
<span>user <code>{{ .Username }}</code></span>
<span>address <code>{{ .RemoteAddr }}</code></span>
{{ with .Stats }}
<span>produced <code>{{ .Delivered }}</code></span>
<span>consumed <code>{{ .Published }}</code></span>
{{ end }}
</small></h2>
{{ with .Subscriptions }}
<ul>
{{ range $topic, $qos := . }}
<li>QoS {{ $qos }} subscription to <code>{{ $topic }}</code></li>
{{ end }}
</ul>
{{ end }}
</section>
{{ end }}
</body>
</html>
`))

type data struct {
	Sessions []session.ServerSession
}

type sessionsByID []session.ServerSession

func (s sessionsByID) Len() int           { return len(s) }
func (s sessionsByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sessionsByID) Less(i, j int) bool { return s[i].ID() < s[j].ID() }

// Sessions inspector
func Sessions(s server.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessions := sessionsByID(s.Sessions().All())
		sort.Sort(sessions)
		tmpl.Execute(w, data{
			Sessions: sessions,
		})
	})
}
