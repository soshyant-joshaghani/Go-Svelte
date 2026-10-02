// Package httpserver wires the modules into one http.Handler.
package httpserver

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/cache"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/config"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/jobs"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/openapi"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/apps/sample"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/auth"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/users"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/system"
)

// Server holds the ports the routes depend on. Tests pass in-memory fakes.
type Server struct {
	Config config.Config
	Users  users.Repository
	Notes  sample.Repository
	Cache  cache.Store
	Jobs   jobs.Queue
}

// UserService is the users service the server builds (main uses it to seed the first superuser).
func (s *Server) UserService() *users.Service {
	return users.NewService(s.Users, s.Notes, s.Config.BcryptCost)
}

// Handler returns the full API: routes, docs and CORS.
func (s *Server) Handler() http.Handler {
	cfg := s.Config
	prefix := cfg.APIV1Str
	mux := http.NewServeMux()

	userSvc := s.UserService()
	authn := &auth.Authenticator{Users: userSvc, Secret: cfg.SecretKey, ExpireMinutes: cfg.AccessTokenExpireMinutes}
	noteSvc := sample.NewService(s.Notes, s.Cache)

	system.Routes(mux, prefix, cfg.IsLocal(), userSvc, s.Jobs)
	auth.Routes(mux, prefix, authn)
	users.Routes(mux, prefix, userSvc, authn.Guards())
	sample.Routes(mux, prefix, noteSvc, authn)
	openapi.Routes(mux, cfg.ProjectName, prefix, cfg.IsLocal())

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		apierr.Write(w, apierr.NotFound("Not Found"))
	})
	return s.cors(recoverer(mux))
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				apierr.Write(w, apierr.New(http.StatusInternalServerError, "Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// cors allows BACKEND_CORS_ORIGINS plus FRONTEND_HOST, with credentials, all methods and headers.
func (s *Server) cors(next http.Handler) http.Handler {
	allowed := s.Config.CORSOrigins()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && slices.Contains(allowed, strings.TrimRight(origin, "/")) {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				h.Set("Access-Control-Allow-Methods", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
				if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
					h.Set("Access-Control-Allow-Headers", reqHeaders)
				}
				h.Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
