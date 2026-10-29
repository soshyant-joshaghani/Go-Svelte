# Database

PostgreSQL 18. The schema is the one Fast's Alembic creates (`"user"` and `note`), so one database works under any FoxG backend.

SQL files live in `backend/migrations`. The API applies them in name order when it starts and records each in `schema_migrations`. Every statement is `IF NOT EXISTS`, so the API can start against a database Fast already built. Queries are explicit SQL in `backend/sql/queries/*.sql` (`-- name: X` blocks), run through pgx by the repositories; there is no code generator.

Adminer: http://adminer.localhost, server `db`, port 5432, credentials from `.env`. From the host or an IDE use `localhost:5432`.

Postgres 18 declares `VOLUME /var/lib/postgresql`. Compose mounts `db-data` there and sets `PGDATA` to `/var/lib/postgresql/18/docker`; mounting the old `/var/lib/postgresql/data` path leaves an anonymous volume. After a Postgres volume change: `dev purge infra`, then `dev run infra`.
