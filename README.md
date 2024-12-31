# Chirpy API

A RESTful API for a social media platform inspired by Twitter, built with Go and PostgreSQL.

## Table of Contents

- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [Environment Variables](#environment-variables)
- [Security](#security)
- [Testing](#testing)
- [Deployment](#deployment)
- [Authors](#authors)

## Getting Started

### Prerequisites

- Go 1.17+
- PostgreSQL 13+
- Stripe account (for payment processing)

### Installation

1. Clone the repository: `git clone https://github.com/Fepozopo/chirpy.git`
2. Install Go: `go install`
3. Install PostgreSQL: `brew install postgresql` (on macOS)
4. Create a PostgreSQL database: `createdb chirpy`
5. Initialize the database schema: `goose postgres 'DB_URL' up`
6. Set environment variables (see below)
7. Run the API: `go run .`

### Environment Variables

- `DB_URL`: the PostgreSQL connection string (e.g. `postgres://user:password@localhost/chirpy?sslmode=disable`)
- `TOKEN_SECRET`: a secret key for generating JSON Web Tokens (e.g. `my_secret_key`)
- `STRIPE_KEY`: your Stripe API key (e.g. `sk_test_...`)

## API Endpoints

### Users

- `POST /api/users`: create a new user
- `GET /api/users/{id}`: retrieve a user by ID
- `PUT /api/users/{id}`: update a user
- `DELETE /api/users/{id}`: delete a user

### Chirps

- `POST /api/chirps`: create a new chirp
- `GET /api/chirps`: retrieve all chirps
- `GET /api/chirps/{id}`: retrieve a chirp by ID
- `DELETE /api/chirps/{id}`: delete a chirp

### Authentication

- `POST /api/login`: authenticate a user and generate a JSON Web Token
- `POST /api/refresh`: refresh a JSON Web Token
- `POST /api/revoke`: revoke a JSON Web Token

### Stripe Webhooks

- `POST /api/polka/webhooks`: handle Stripe webhooks (e.g. user.upgraded)

## Database Schema

The database schema is defined in [sql/schema](sql/schema). It consists of the following tables:

- `users`: stores user information (e.g. email, hashed password)
- `chirps`: stores chirp information (e.g. body, user ID)
- `refresh_tokens`: stores refresh tokens (e.g. token, user ID, expiration date)

## Security

The API uses JSON Web Tokens for authentication and authorization. The `TOKEN_SECRET` environment variable is used to generate and verify tokens.

## Testing

The API includes unit tests and integration tests. To run the tests, execute `go test ./...`.

## Authors

- [Fepozopo](https://github.com/Fepozopo)
