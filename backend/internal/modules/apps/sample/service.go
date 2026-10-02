// Notes CRUD with a Redis read cache (soft-degrading).
//
// Keys: sample:notes:v1:list:<owner> (120 s) and sample:notes:v1:note:<owner>:<id> (300 s).
package sample

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/cache"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/httpx"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/users"
)

const (
	CachePrefix = "sample:notes:v1:"
	TTLList     = 120 * time.Second
	TTLNote     = 300 * time.Second
)

func ListKey(owner string) string         { return CachePrefix + "list:" + owner }
func NoteKey(owner, noteID string) string { return CachePrefix + "note:" + owner + ":" + noteID }

// now keeps microseconds: Postgres does, and cached and stored values must agree.
func now() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }

var (
	errNotAllowed = func() error { return apierr.Forbidden("Not allowed to access this note") }
	errTitleEmpty = func() error { return apierr.Validation("Title cannot be empty") }
)

type Service struct {
	notes Repository
	cache cache.Store
}

func NewService(notes Repository, cache cache.Store) *Service {
	return &Service{notes: notes, cache: cache}
}

func internal(err error) error { return apierr.Internal(err) }

func (s *Service) cachePut(ctx context.Context, key string, value any, ttl time.Duration) {
	if data, err := json.Marshal(value); err == nil {
		s.cache.Set(ctx, key, string(data), ttl)
	}
}

func (s *Service) invalidateLists(ctx context.Context, owner string) {
	s.cache.DeletePrefix(ctx, ListKey(owner))
}

// loadOwned reads from Postgres (never the cache) and checks ownership: 404 then 403.
func (s *Service) loadOwned(ctx context.Context, user *users.User, noteID string) (*Note, error) {
	note, err := s.notes.GetByID(ctx, noteID)
	if err != nil {
		return nil, internal(err)
	}
	if note == nil {
		return nil, apierr.NotFound("Note not found")
	}
	if note.OwnerID != user.ID {
		return nil, errNotAllowed()
	}
	return note, nil
}

func (s *Service) List(ctx context.Context, user *users.User) ([]NotePublic, error) {
	key := ListKey(user.ID)
	if raw, ok := s.cache.Get(ctx, key); ok {
		var cached []NotePublic
		if json.Unmarshal([]byte(raw), &cached) == nil && cached != nil {
			return cached, nil
		}
	}
	rows, err := s.notes.ListByOwner(ctx, user.ID)
	if err != nil {
		return nil, internal(err)
	}
	out := make([]NotePublic, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Public())
	}
	s.cachePut(ctx, key, out, TTLList)
	return out, nil
}

func (s *Service) Create(ctx context.Context, user *users.User, in NoteCreate) (NotePublic, error) {
	if err := httpx.CheckLen("title", in.Title, 1, 255); err != nil {
		return NotePublic{}, err
	}
	if err := httpx.CheckLen("content", in.Content, 0, 10000); err != nil {
		return NotePublic{}, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return NotePublic{}, errTitleEmpty()
	}
	stamp := now()
	saved, err := s.notes.Create(ctx, Note{
		ID: uuid.NewString(), Title: title, Content: strings.TrimSpace(in.Content),
		OwnerID: user.ID, CreatedAt: stamp, UpdatedAt: stamp,
	})
	if err != nil {
		return NotePublic{}, internal(err)
	}
	public := saved.Public()
	s.invalidateLists(ctx, user.ID)
	s.cachePut(ctx, NoteKey(user.ID, saved.ID), public, TTLNote)
	return public, nil
}

func (s *Service) Get(ctx context.Context, user *users.User, noteID string) (NotePublic, error) {
	key := NoteKey(user.ID, noteID)
	if raw, ok := s.cache.Get(ctx, key); ok {
		var cached NotePublic
		if json.Unmarshal([]byte(raw), &cached) == nil && cached.ID != "" {
			if cached.OwnerID != user.ID {
				return NotePublic{}, errNotAllowed()
			}
			return cached, nil
		}
	}
	note, err := s.loadOwned(ctx, user, noteID)
	if err != nil {
		return NotePublic{}, err
	}
	public := note.Public()
	s.cachePut(ctx, key, public, TTLNote)
	return public, nil
}

func (s *Service) Update(ctx context.Context, user *users.User, noteID string, in NoteUpdate) (NotePublic, error) {
	if in.Title != nil {
		if err := httpx.CheckLen("title", *in.Title, 1, 255); err != nil {
			return NotePublic{}, err
		}
	}
	if in.Content != nil {
		if err := httpx.CheckLen("content", *in.Content, 0, 10000); err != nil {
			return NotePublic{}, err
		}
	}
	note, err := s.loadOwned(ctx, user, noteID)
	if err != nil {
		return NotePublic{}, err
	}
	changed := false
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return NotePublic{}, errTitleEmpty()
		}
		note.Title, changed = title, true
	}
	if in.Content != nil {
		note.Content, changed = strings.TrimSpace(*in.Content), true
	}
	if changed {
		note.UpdatedAt = now()
		saved, err := s.notes.Update(ctx, *note)
		if err != nil {
			return NotePublic{}, internal(err)
		}
		note = &saved
	}
	public := note.Public()
	s.invalidateLists(ctx, user.ID)
	s.cachePut(ctx, NoteKey(user.ID, noteID), public, TTLNote)
	return public, nil
}

func (s *Service) Delete(ctx context.Context, user *users.User, noteID string) error {
	note, err := s.loadOwned(ctx, user, noteID)
	if err != nil {
		return err
	}
	if err := s.notes.Delete(ctx, note.ID); err != nil {
		return internal(err)
	}
	s.cache.Delete(ctx, NoteKey(user.ID, noteID))
	s.invalidateLists(ctx, user.ID)
	return nil
}
