package sample

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlq "github.com/soshyant-joshaghani/go-svelte/sql"
)

// Repository holds the queries for table note. Missing rows are (nil, nil).
type Repository interface {
	GetByID(ctx context.Context, id string) (*Note, error)
	ListByOwner(ctx context.Context, ownerID string) ([]Note, error)
	Create(ctx context.Context, note Note) (Note, error)
	Update(ctx context.Context, note Note) (Note, error)
	Delete(ctx context.Context, id string) error
	// DeleteByOwner also satisfies users.NoteCleaner.
	DeleteByOwner(ctx context.Context, ownerID string) error
}

type PgRepository struct{ pool *pgxpool.Pool }

func NewPgRepository(pool *pgxpool.Pool) *PgRepository { return &PgRepository{pool: pool} }

func scanNote(row pgx.Row) (Note, error) {
	var n Note
	err := row.Scan(&n.ID, &n.Title, &n.Content, &n.OwnerID, &n.CreatedAt, &n.UpdatedAt)
	n.CreatedAt, n.UpdatedAt = n.CreatedAt.UTC(), n.UpdatedAt.UTC()
	return n, err
}

func (r *PgRepository) GetByID(ctx context.Context, id string) (*Note, error) {
	n, err := scanNote(r.pool.QueryRow(ctx, sqlq.Get("GetNoteByID"), id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *PgRepository) ListByOwner(ctx context.Context, ownerID string) ([]Note, error) {
	rows, err := r.pool.Query(ctx, sqlq.Get("ListNotesByOwner"), ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Note{}
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *PgRepository) Create(ctx context.Context, n Note) (Note, error) {
	_, err := r.pool.Exec(ctx, sqlq.Get("CreateNote"), n.ID, n.Title, n.Content, n.OwnerID, n.CreatedAt, n.UpdatedAt)
	return n, err
}

func (r *PgRepository) Update(ctx context.Context, n Note) (Note, error) {
	_, err := r.pool.Exec(ctx, sqlq.Get("UpdateNote"), n.ID, n.Title, n.Content, n.UpdatedAt)
	return n, err
}

func (r *PgRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, sqlq.Get("DeleteNote"), id)
	return err
}

func (r *PgRepository) DeleteByOwner(ctx context.Context, ownerID string) error {
	_, err := r.pool.Exec(ctx, sqlq.Get("DeleteNotesByOwner"), ownerID)
	return err
}
