// Package sql loads the named queries in sql/queries/*.sql.
//
// Each query starts with a "-- name: <Name>" comment and ends at the next one.
// Repositories fetch the text with Get and run it through pgx.
package sql

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed queries/*.sql
var files embed.FS

var queries = load()

func load() map[string]string {
	out := map[string]string{}
	entries, err := files.ReadDir("queries")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		data, err := files.ReadFile("queries/" + entry.Name())
		if err != nil {
			panic(err)
		}
		name := ""
		var body []string
		flush := func() {
			if name != "" {
				out[name] = strings.TrimSpace(strings.Join(body, "\n"))
			}
		}
		for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
			if rest, ok := strings.CutPrefix(line, "-- name:"); ok {
				flush()
				name, body = strings.TrimSpace(rest), nil
				continue
			}
			body = append(body, line)
		}
		flush()
	}
	return out
}

// Get returns the query text. It panics on an unknown name (a programming error).
func Get(name string) string {
	text, ok := queries[name]
	if !ok {
		panic(fmt.Sprintf("sql: unknown query %q", name))
	}
	return text
}
