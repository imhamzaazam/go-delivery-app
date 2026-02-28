
# Go Web API Boilerplate

Just writing some golang with things that I like to see in a Web API

## Features

- Hexagonal Architecture (kinda overengineering but ok. Also, just wrote like this to see how it goes)
- Simple routing with chi
- Centralized encoding and decoding
- Centralized error handling
- Versioned HTTP Handler
- SQL type safety with SQLC
- Migrations with golang migrate
- JWT authentication with symmetric secret key
- Access and Refresh Tokens
- Tests that uses Testcontainers instead of mocks
- Testing scripts that uses cURL and jq (f* Postman)

## JWT setup

Set:

- `TOKEN_SYMMETRIC_KEY=super-secret-jwt-key-32-characters`

## Required dependencies

- jq
- golang-migrate
- docker
- sqlc
