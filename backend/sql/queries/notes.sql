-- name: GetNoteByID
SELECT id::text, title, content, owner_id::text, created_at, updated_at
FROM note WHERE id = $1::uuid;

-- name: ListNotesByOwner
SELECT id::text, title, content, owner_id::text, created_at, updated_at
FROM note WHERE owner_id = $1::uuid ORDER BY updated_at DESC;

-- name: CreateNote
INSERT INTO note (id, title, content, owner_id, created_at, updated_at)
VALUES ($1::uuid, $2, $3, $4::uuid, $5, $6);

-- name: UpdateNote
UPDATE note SET title = $2, content = $3, updated_at = $4 WHERE id = $1::uuid;

-- name: DeleteNote
DELETE FROM note WHERE id = $1::uuid;

-- name: DeleteNotesByOwner
DELETE FROM note WHERE owner_id = $1::uuid;
