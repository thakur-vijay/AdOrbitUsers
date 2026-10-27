# AdOrbitUsers

AdOrbitUsers is the user-management microservice for AdOrbit. It is written in Go, uses PostgreSQL for persistence, and delegates access-token validation to the AdOrbit authentication service.

## Requirements

- Go 1.27.1
- Docker with Docker Compose
- PostgreSQL (or use the Compose service below)
- The AdOrbit authentication service reachable at `127.0.0.1:9090`

## Configuration

The service loads a `.env` file from the working directory when present. Configure:

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `APP_PORT` | No | `8080` | HTTP server port |
| `DATABASE_URL` | Yes | — | PostgreSQL connection URL |
| `POSTGRES_USER` | For Compose | — | PostgreSQL user |
| `POSTGRES_PASSWORD` | For Compose | — | PostgreSQL password |
| `POSTGRES_DB` | For Compose | — | PostgreSQL database name |

For local development, create a `.env` file such as:

```dotenv
APP_PORT=8080
DATABASE_URL=YOUR_POSTGRES_CONNECTION_URL
POSTGRES_USER=adorbit
POSTGRES_PASSWORD=change-me
POSTGRES_DB=adorbit_users
```

Use your own database credentials and do not commit secrets.

## Run locally

Start the PostgreSQL container:

```sh
docker compose up -d postgres
```

Then start the service from the repository root:

```sh
go run ./cmd/server
```

On startup, the service connects to PostgreSQL and applies database migrations. The authentication service must also be running at `127.0.0.1:9090`; user routes require authentication.

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/users` | Create a user |
| `GET` | `/api/v1/users` | Fetch users |
| `GET` | `/api/v1/users/{id}` | Fetch a user by ID |
| `PUT` | `/api/v1/users/{id}` | Update a user |
| `DELETE` | `/api/v1/users/{id}` | Delete a user |

All `/api/v1/users` routes are protected by the authentication middleware. The user handler implementations are in progress; currently, the create handler validates its request, while the fetch, update, and delete handlers are placeholders.

Check service health with:

```sh
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## Database

The service applies SQL migrations from `internal/database/migrations` at startup. The initial migration creates the `users` table with a UUID ID, full name, unique email, role, and creation/update timestamps.
