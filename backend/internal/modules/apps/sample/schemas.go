package sample

import "time"

// Note is the row of table note.
type Note struct {
	ID        string
	Title     string
	Content   string
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoteCreate struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type NoteUpdate struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

type NotePublic struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (n Note) Public() NotePublic {
	return NotePublic{
		ID: n.ID, Title: n.Title, Content: n.Content, OwnerID: n.OwnerID,
		CreatedAt: n.CreatedAt.UTC(), UpdatedAt: n.UpdatedAt.UTC(),
	}
}
