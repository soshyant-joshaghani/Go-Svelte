// Package openapi serves the hand-written OpenAPI document plus Swagger UI
// (/docs) and Scalar (/sdoc).
package openapi

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
)

const scheme = "OAuth2PasswordBearer"

type obj = map[string]any

func ref(name string) obj { return obj{"$ref": "#/components/schemas/" + name} }

func content(schema obj) obj { return obj{"application/json": obj{"schema": schema}} }

type op struct {
	path, method, tag, summary string
	secured, form              bool
	request, response          obj
	status                     string
}

func add(paths obj, o op) {
	if o.status == "" {
		o.status = "200"
	}
	operation := obj{"tags": []string{o.tag}, "summary": o.summary}
	if strings.Contains(o.path, "{id}") {
		operation["parameters"] = []obj{{
			"name": "id", "in": "path", "required": true,
			"schema": obj{"type": "string", "format": "uuid"},
		}}
	}
	if o.secured {
		operation["security"] = []obj{{scheme: []string{}}}
	}
	if o.request != nil {
		operation["requestBody"] = obj{"required": true, "content": content(o.request)}
	}
	if o.form {
		operation["requestBody"] = obj{"required": true, "content": obj{
			"application/x-www-form-urlencoded": obj{"schema": obj{
				"type": "object", "required": []string{"username", "password"},
				"properties": obj{
					"username": obj{"type": "string"},
					"password": obj{"type": "string", "format": "password"},
				},
			}},
		}}
	}
	ok := obj{"description": "Successful Response"}
	if o.response != nil {
		ok["content"] = content(o.response)
	}
	operation["responses"] = obj{
		o.status: ok,
		"4XX":    obj{"description": "Error", "content": content(ref("Detail"))},
	}
	entry, exists := paths[o.path].(obj)
	if !exists {
		entry = obj{}
		paths[o.path] = entry
	}
	entry[o.method] = operation
}

func components(tokenURL string) obj {
	str := func(extra obj) obj {
		out := obj{"type": "string"}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	nullable := obj{"type": []string{"string", "null"}}
	password := obj{"type": "string", "minLength": 8, "maxLength": 128}
	title := obj{"type": "string", "minLength": 1, "maxLength": 255}
	return obj{
		"securitySchemes": obj{scheme: obj{
			"type":  "oauth2",
			"flows": obj{"password": obj{"tokenUrl": tokenURL, "scopes": obj{}}},
		}},
		"schemas": obj{
			"Detail":  obj{"type": "object", "properties": obj{"detail": obj{"type": "string"}}},
			"Message": obj{"type": "object", "required": []string{"message"}, "properties": obj{"message": obj{"type": "string"}}},
			"Token": obj{"type": "object", "required": []string{"access_token", "token_type"}, "properties": obj{
				"access_token": obj{"type": "string"}, "token_type": obj{"type": "string"},
			}},
			"UserPublic": obj{"type": "object", "required": []string{"id", "email", "is_active", "is_superuser"}, "properties": obj{
				"id": str(obj{"format": "uuid"}), "email": obj{"type": "string"},
				"is_active": obj{"type": "boolean"}, "is_superuser": obj{"type": "boolean"}, "full_name": nullable,
			}},
			"UsersPublic": obj{"type": "object", "required": []string{"data", "count"}, "properties": obj{
				"data": obj{"type": "array", "items": ref("UserPublic")}, "count": obj{"type": "integer"},
			}},
			"UserCreate": obj{"type": "object", "required": []string{"email", "password"}, "properties": obj{
				"email": obj{"type": "string", "maxLength": 255}, "password": password,
				"is_active": obj{"type": "boolean", "default": true}, "is_superuser": obj{"type": "boolean", "default": false},
				"full_name": obj{"type": []string{"string", "null"}, "maxLength": 255},
			}},
			"UserUpdate": obj{"type": "object", "properties": obj{
				"email": obj{"type": "string", "maxLength": 255}, "password": password,
				"is_active": obj{"type": "boolean"}, "is_superuser": obj{"type": "boolean"},
				"full_name": obj{"type": []string{"string", "null"}, "maxLength": 255},
			}},
			"PrivateUserCreate": obj{"type": "object", "required": []string{"email", "password"}, "properties": obj{
				"email": obj{"type": "string", "maxLength": 255}, "password": password,
				"full_name": obj{"type": []string{"string", "null"}, "maxLength": 255},
			}},
			"NoteCreate": obj{"type": "object", "required": []string{"title"}, "properties": obj{
				"title": title, "content": obj{"type": "string", "maxLength": 10000, "default": ""},
			}},
			"NoteUpdate": obj{"type": "object", "properties": obj{
				"title": title, "content": obj{"type": "string", "maxLength": 10000},
			}},
			"NotePublic": obj{"type": "object", "required": []string{"id", "title", "content", "owner_id", "created_at", "updated_at"}, "properties": obj{
				"id": str(obj{"format": "uuid"}), "title": obj{"type": "string"}, "content": obj{"type": "string"},
				"owner_id":   str(obj{"format": "uuid"}),
				"created_at": str(obj{"format": "date-time"}), "updated_at": str(obj{"format": "date-time"}),
			}},
		},
	}
}

// BuildSpec builds the OpenAPI 3.1 document for the routes this server exposes.
func BuildSpec(projectName, apiPrefix string, local bool) obj {
	const (
		auth    = "[BASE] Auth"
		users   = "[SUPERADMIN] Core - User Management"
		sample  = "[APPS] Sample"
		utils   = "[SYSTEM] System - Utils"
		private = "[SYSTEM] System - Private"
	)
	paths := obj{}
	add(paths, op{path: "/utils/health-check", method: "get", tag: utils, summary: "Health Check"})
	if local {
		add(paths, op{path: "/private/ping", method: "get", tag: private, summary: "Private Ping", response: ref("Message")})
		add(paths, op{path: "/private/users", method: "post", tag: private, summary: "Create User (local)", request: ref("PrivateUserCreate"), response: ref("UserPublic")})
		add(paths, op{path: "/private/jobs/ping", method: "post", tag: private, summary: "Enqueue Ping Job"})
	}
	add(paths, op{path: "/base/login/access-token", method: "post", tag: auth, summary: "Login Access Token", form: true, response: ref("Token")})
	add(paths, op{path: "/base/login/me", method: "get", tag: auth, summary: "Read Users Me", secured: true, response: ref("UserPublic")})

	add(paths, op{path: "/base/users/admin", method: "get", tag: users, summary: "Read Users", secured: true, response: ref("UsersPublic")})
	add(paths, op{path: "/base/users/admin", method: "post", tag: users, summary: "Create User", secured: true, request: ref("UserCreate"), response: ref("UserPublic")})
	add(paths, op{path: "/base/users/{id}/admin", method: "get", tag: users, summary: "Read User By Id", secured: true, response: ref("UserPublic")})
	add(paths, op{path: "/base/users/{id}/admin", method: "patch", tag: users, summary: "Update User", secured: true, request: ref("UserUpdate"), response: ref("UserPublic")})
	add(paths, op{path: "/base/users/{id}/admin", method: "delete", tag: users, summary: "Delete User", secured: true, response: ref("Message")})

	add(paths, op{path: "/sample", method: "get", tag: sample, summary: "Sample Root", response: ref("Message")})
	add(paths, op{path: "/sample/notes", method: "get", tag: sample, summary: "List Notes", secured: true, response: obj{"type": "array", "items": ref("NotePublic")}})
	add(paths, op{path: "/sample/notes", method: "post", tag: sample, summary: "Create Note", secured: true, request: ref("NoteCreate"), status: "201", response: ref("NotePublic")})
	add(paths, op{path: "/sample/notes/{id}", method: "get", tag: sample, summary: "Read Note", secured: true, response: ref("NotePublic")})
	add(paths, op{path: "/sample/notes/{id}", method: "patch", tag: sample, summary: "Update Note", secured: true, request: ref("NoteUpdate"), response: ref("NotePublic")})
	add(paths, op{path: "/sample/notes/{id}", method: "delete", tag: sample, summary: "Delete Note", secured: true, status: "204"})

	return obj{
		"openapi":    "3.1.0",
		"info":       obj{"title": projectName, "version": "0.1.0"},
		"servers":    []obj{{"url": apiPrefix}},
		"paths":      paths,
		"components": components(apiPrefix + "/base/login/access-token"),
	}
}

func swaggerHTML(title, specURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>%s - Swagger UI</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
window.ui = SwaggerUIBundle({ url: %q, dom_id: "#swagger-ui", persistAuthorization: true });
</script>
</body>
</html>`, html.EscapeString(title), specURL)
}

func scalarHTML(title, specURL string) string {
	config, _ := json.Marshal(obj{
		"theme":          "elysiajs",
		"layout":         "modern",
		"persistAuth":    true,
		"authentication": obj{"preferredSecurityScheme": scheme},
	})
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>%s - Scalar</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
<script id="api-reference" data-url=%q data-configuration='%s'></script>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`, html.EscapeString(title), specURL, config)
}

// Routes registers /docs, /sdoc and <apiPrefix>/openapi.json.
func Routes(mux *http.ServeMux, projectName, apiPrefix string, local bool) {
	specURL := apiPrefix + "/openapi.json"
	spec, _ := json.Marshal(BuildSpec(projectName, apiPrefix, local))
	swagger, scalar := swaggerHTML(projectName, specURL), scalarHTML(projectName, specURL)

	page := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(body))
		}
	}
	mux.HandleFunc("GET /docs", page(swagger))
	mux.HandleFunc("GET /sdoc", page(scalar))
	mux.HandleFunc("GET "+specURL, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(spec)
	})
}
