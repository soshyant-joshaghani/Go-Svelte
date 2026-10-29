package users

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/apierr"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/httpx"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/security"
)

const (
	duplicateEmail = "The user with this email already exists in the system."
	noPrivileges   = "The user doesn't have enough privileges"
)

// NoteCleaner deletes a user's notes before the user row goes (the sample
// module's repository implements it).
type NoteCleaner interface {
	DeleteByOwner(ctx context.Context, ownerID string) error
}

type Service struct {
	repo       Repository
	notes      NoteCleaner
	bcryptCost int
}

func NewService(repo Repository, notes NoteCleaner, bcryptCost int) *Service {
	return &Service{repo: repo, notes: notes, bcryptCost: bcryptCost}
}

func (s *Service) hash(password string) (string, error) {
	hashed, err := security.HashPassword(password, s.bcryptCost)
	if err != nil {
		return "", apierr.Internal(err)
	}
	return hashed, nil
}

func internal(err error) error {
	var api *apierr.Error
	if errors.As(err, &api) {
		return err
	}
	return apierr.Internal(err)
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	return user, nil
}

// Authenticate is the login check. Wrong email/password and inactive users are 400.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, internal(err)
	}
	if user == nil || !security.VerifyPassword(password, user.HashedPassword) {
		return nil, apierr.BadRequest("Incorrect email or password")
	}
	if !user.IsActive {
		return nil, apierr.BadRequest("Inactive user")
	}
	return user, nil
}

func (s *Service) List(ctx context.Context, skip, limit int) (UsersPublic, error) {
	if skip < 0 {
		skip = 0
	}
	if limit < 0 {
		limit = 0
	}
	count, err := s.repo.Count(ctx)
	if err != nil {
		return UsersPublic{}, internal(err)
	}
	rows, err := s.repo.List(ctx, skip, limit)
	if err != nil {
		return UsersPublic{}, internal(err)
	}
	out := UsersPublic{Data: make([]UserPublic, 0, len(rows)), Count: count}
	for _, row := range rows {
		out.Data = append(out.Data, row.Public())
	}
	return out, nil
}

func (in UserCreate) validate() error {
	if err := httpx.CheckEmail("email", in.Email); err != nil {
		return err
	}
	if err := httpx.CheckLen("password", in.Password, 8, 128); err != nil {
		return err
	}
	if in.FullName != nil {
		return httpx.CheckLen("full_name", *in.FullName, 0, 255)
	}
	return nil
}

func (in UserUpdate) validate() error {
	if in.Email != nil {
		if err := httpx.CheckEmail("email", *in.Email); err != nil {
			return err
		}
	}
	if in.Password != nil {
		if err := httpx.CheckLen("password", *in.Password, 8, 128); err != nil {
			return err
		}
	}
	if in.FullName.Value != nil {
		return httpx.CheckLen("full_name", *in.FullName.Value, 0, 255)
	}
	return nil
}

func (s *Service) Create(ctx context.Context, in UserCreate) (*User, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, internal(err)
	}
	if existing != nil {
		return nil, apierr.BadRequest(duplicateEmail)
	}
	hashed, err := s.hash(in.Password)
	if err != nil {
		return nil, err
	}
	user := User{
		ID:             uuid.NewString(),
		Email:          in.Email,
		IsActive:       in.IsActive == nil || *in.IsActive,
		IsSuperuser:    in.IsSuperuser != nil && *in.IsSuperuser,
		FullName:       in.FullName,
		HashedPassword: hashed,
	}
	saved, err := s.repo.Create(ctx, user)
	if errors.Is(err, ErrDuplicateEmail) {
		return nil, apierr.BadRequest(duplicateEmail)
	}
	if err != nil {
		return nil, internal(err)
	}
	return &saved, nil
}

// GetFor lets a user read itself; anyone else needs a superuser.
func (s *Service) GetFor(ctx context.Context, current *User, id string) (*User, error) {
	found, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	if found != nil && found.ID == current.ID {
		return found, nil
	}
	if !current.IsSuperuser {
		return nil, apierr.Forbidden(noPrivileges)
	}
	if found == nil {
		return nil, apierr.NotFound("User not found")
	}
	return found, nil
}

func (s *Service) Update(ctx context.Context, id string, in UserUpdate) (*User, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, internal(err)
	}
	if user == nil {
		return nil, apierr.NotFound("The user with this id does not exist in the system")
	}
	if in.Email != nil {
		other, err := s.repo.GetByEmail(ctx, *in.Email)
		if err != nil {
			return nil, internal(err)
		}
		if other != nil && other.ID != id {
			return nil, apierr.Conflict("User with this email already exists")
		}
		user.Email = *in.Email
	}
	if in.Password != nil {
		if user.HashedPassword, err = s.hash(*in.Password); err != nil {
			return nil, err
		}
	}
	if in.FullName.Set {
		user.FullName = in.FullName.Value
	}
	if in.IsActive != nil {
		user.IsActive = *in.IsActive
	}
	if in.IsSuperuser != nil {
		user.IsSuperuser = *in.IsSuperuser
	}
	saved, err := s.repo.Update(ctx, *user)
	if errors.Is(err, ErrDuplicateEmail) {
		return nil, apierr.Conflict("User with this email already exists")
	}
	if err != nil {
		return nil, internal(err)
	}
	return &saved, nil
}

// Delete removes the user's notes first, then the user.
func (s *Service) Delete(ctx context.Context, current *User, id string) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return internal(err)
	}
	if user == nil {
		return apierr.NotFound("User not found")
	}
	if user.ID == current.ID {
		return apierr.Forbidden("Super users are not allowed to delete themselves")
	}
	if err := s.notes.DeleteByOwner(ctx, id); err != nil {
		return internal(err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return internal(err)
	}
	return nil
}

// EnsureFirstSuperuser creates the configured superuser when it is missing.
func (s *Service) EnsureFirstSuperuser(ctx context.Context, email, password string) error {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	hashed, err := security.HashPassword(password, s.bcryptCost)
	if err != nil {
		return err
	}
	if _, err := s.repo.Create(ctx, User{
		ID: uuid.NewString(), Email: email, IsActive: true, IsSuperuser: true, HashedPassword: hashed,
	}); err != nil && !errors.Is(err, ErrDuplicateEmail) {
		return err
	}
	log.Printf("created first superuser %s", email)
	return nil
}

// Privileges is the 403 shared with the guards.
func Privileges() error { return apierr.Forbidden(noPrivileges) }
