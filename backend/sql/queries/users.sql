-- name: GetUserByID
SELECT id::text, email, is_active, is_superuser, full_name, hashed_password
FROM "user" WHERE id = $1::uuid;

-- name: GetUserByEmail
SELECT id::text, email, is_active, is_superuser, full_name, hashed_password
FROM "user" WHERE email = $1;

-- name: CountUsers
SELECT count(*) FROM "user";

-- name: ListUsers
SELECT id::text, email, is_active, is_superuser, full_name, hashed_password
FROM "user" ORDER BY email OFFSET $1 LIMIT $2;

-- name: CreateUser
INSERT INTO "user" (id, email, is_active, is_superuser, full_name, hashed_password)
VALUES ($1::uuid, $2, $3, $4, $5, $6);

-- name: UpdateUser
UPDATE "user"
SET email = $2, is_active = $3, is_superuser = $4, full_name = $5, hashed_password = $6
WHERE id = $1::uuid;

-- name: DeleteUser
DELETE FROM "user" WHERE id = $1::uuid;
