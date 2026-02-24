# Backend (API)

Quick reference for running the backend and applying database migrations inside the Docker Compose environment.

## Run the stack

- Uses Docker Compose in the project root. To build and start API + dependencies:

```bash
docker compose up --build
```

The API listens on the port defined by env / config (default `8080`).

## Environment

Common env vars (also loaded from `internal/config` defaults):

- `DATABASE_URL` or `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`
- `JWT_SECRET`
- `REDIS_URL` / redis config
- `RABBITMQ_*` for RabbitMQ
- `SERVER_PORT` (default `8080`)

## Apply migrations (from host)

If your Compose service for Postgres is named `postgres` you can apply SQL migration files with:

```bash
# single file
docker compose exec -T postgres psql -U sociomile -d sociomile_db < backend/migrations/postgres_init.sql

# apply all .sql files in migrations (bash)
for f in backend/migrations/*.sql; do 
	echo "Applying $f"; 
	docker compose exec -T postgres psql -U sociomile -d sociomile_db < "$f"; 
done
```

If your postgres container has a different service name update `postgres` accordingly.

## Useful docker commands

```bash
# run a one-off shell inside the api container
docker compose exec api sh

# run psql from host (if psql installed locally)
PGPASSWORD=password psql -h localhost -p 5432 -U sociomile -d sociomile_db
```

## API Endpoints (v1)

Base path: `/api/v1`

Public
- `POST /auth/login` — login (returns JWT)
- `POST /auth/register` — register
- `POST /channel/webhook` — channel simulator/webhook receiver (creates conversation/message)

Protected (require `Authorization: Bearer <token>`)

Channels
- `GET /channels` — list channels
- `GET /channels/:id` — get channel
- `POST /channels` — create channel
- `PUT /channels/:id` — update channel
- `DELETE /channels/:id` — delete channel

Conversations
- `GET /conversations` — list conversations
- `GET /conversations/:id` — get conversation + messages
- `POST /conversations` — create conversation
- `PUT /conversations/:id` — update conversation
- `DELETE /conversations/:id` — delete conversation
- `POST /conversations/:id/messages` — send message (alias for messages endpoint)
- `POST /conversations/:id/assign` — assign conversation
- `POST /conversations/:id/close` — close conversation

Messages
- `DELETE /messages/:id` — delete message

Tickets
- `GET /tickets` — list tickets
- `GET /tickets/:id` — get ticket
- `POST /conversations/:id/escalate` — escalate conversation to ticket
- `POST /tickets` — create ticket
- `PUT /tickets/:id` — update ticket
- `DELETE /tickets/:id` — delete ticket

Admin (requires admin role)
- `PUT /tickets/:id/status` — update ticket status
- `GET /users`, `POST /users`, `PUT /users/:id`, `DELETE /users/:id`

## Health & Websocket

- `GET /health` — healthcheck
- `GET /ws` — websocket upgrade endpoint (for realtime)

## Notes

- When testing from the host shell, avoid shell quoting pitfalls by using files for JSON bodies or using small Python scripts to POST requests (the repository contains examples used during development).
- If you encounter issues with missing demo users, re-run the seeders listed in `backend/migrations`.
