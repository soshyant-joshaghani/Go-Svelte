package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlq "github.com/soshyant-joshaghani/go-svelte/sql"
)

// ErrDuplicateEmail is returned when the unique index on email is hit.
var ErrDuplicateEmail = errors.New("duplicate email")

// Repository holds the queries for table "user". Missing rows are (nil, nil).
type Repository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, skip, limit int) ([]User, error)
	Create(ctx context.Context, user User) (User, error)
	Update(ctx context.Context, user User) (User, error)
	Delete(ctx context.Context, id string) error
}

type PgRepository struct{ pool *pgxpool.Pool }

func NewPgRepository(pool *pgxpool.Pool) *PgRepository { return &PgRepository{pool: pool} }

func scan(row pgx.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.IsActive, &u.IsSuperuser, &u.FullName, &u.HashedPassword)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func isUnique(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "23505"
}

func (r *PgRepository) GetByID(ctx context.Context, id string) (*User, error) {
	return scan(r.pool.QueryRow(ctx, sqlq.Get("GetUserByID"), id))
}

func (r *PgRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return scan(r.pool.QueryRow(ctx, sqlq.Get("GetUserByEmail"), email))
}

func (r *PgRepository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, sqlq.Get("CountUsers")).Scan(&n)
	return n, err
}

func (r *PgRepository) List(ctx context.Context, skip, limit int) ([]User, error) {
	rows, err := r.pool.Query(ctx, sqlq.Get("ListUsers"), skip, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.IsActive, &u.IsSuperuser, &u.FullName, &u.HashedPassword); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *PgRepository) Create(ctx context.Context, u User) (User, error) {
	_, err := r.pool.Exec(ctx, sqlq.Get("CreateUser"), u.ID, u.Email, u.IsActive, u.IsSuperuser, u.FullName, u.HashedPassword)
	if isUnique(err) {
		return User{}, ErrDuplicateEmail
	}
	return u, err
}

func (r *PgRepository) Update(ctx context.Context, u User) (User, error) {
	_, err := r.pool.Exec(ctx, sqlq.Get("UpdateUser"), u.ID, u.Email, u.IsActive, u.IsSuperuser, u.FullName, u.HashedPassword)
	if isUnique(err) {
		return User{}, ErrDuplicateEmail
	}
	return u, err
}

func (r *PgRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, sqlq.Get("DeleteUser"), id)
	return err
}
