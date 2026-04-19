# DeepIdle Game Server

Go-based backend server for the AI-driven idle game DeepIdle.

## Overview

A server-authoritative idle game where:
- Server simulates all gameplay logic
- Players issue high-level intents (e.g., `chop_tree`)
- Actions execute as long-running jobs
- AI bridges human intent and game actions

## Prerequisites

- Go 1.22+
- MongoDB (local or Atlas)
- .env file with configuration

## Installation

```bash
cd server

# Install dependencies
go mod download

# Create .env file
cp .env.example .env
# Edit .env with your MongoDB URI and JWT secret
```

### Environment Variables

Create a `.env` file:

```env
PORT=3000
MONGO_URI=mongodb+srv://username:password@cluster.mongodb.net/
DB_NAME=deepidle
JWT_SECRET=your-secret-key-here
```

## Running

```bash
# Development
go run main.go

# Production
go build -o server
./server
```

Server starts on port 3000 (or PORT env var).

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/auth/signup | Create new account |
| POST | /api/auth/signin | Login, returns JWT |

### Character
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/character/detail | Get character info |
| POST | /api/character/action | Set action |
| POST | /api/character/claim | Claim resources |
| PATCH | /api/character/name | Update name |

### Inventory
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/inventory | Get inventory |
| GET | /api/inventory/upgrade-options | Get upgrade costs |
| POST | /api/inventory/upgrade | Upgrade item |

### Storage (Global, shared)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/storage | Get global storage |
| POST | /api/storage/deposit | Deposit item |
| POST | /api/storage/withdraw | Withdraw item |

## Game Actions

| Action | Tool Required | Produces |
|--------|---------------|----------|
| cutting_wood | wooden_axe | wood |
| mining_rocks | wooden_pickaxe | stone |
| fighting_monsters | wooden_sword | gold |

## Upgrade Costs

| Item | Level N→N+1 |
|------|-------------|
| wooden_axe | wood: N×5, gold: N×10 |
| wooden_pickaxe | stone: N×5, wood: N×5, gold: N×10 |
| wooden_sword | wood: N×5, gold: N×10 |

## Project Structure

```
server/
├── main.go              # Entry point
├── handlers/            # HTTP handlers
│   ├── auth.go
│   ├── character.go
│   ├── inventory.go
│   └── storage.go
├── routes/              # Route registration
├── models/              # Data models
├── database/            # MongoDB connection
├── middleware/          # Auth middleware
├── state/               # Game state
├── go.mod
├── go.sum
├── .env                 # Credentials (never commit)
└── .gitignore
```

## Database

Uses MongoDB with collections:
- `characters` - Player accounts and state
- `inventories` - Player inventory items
- `configs` - Server configuration
- `global_storage` - Shared storage across all players

## Security

- JWT-based authentication
- Server-side validation only
- No trust in client input
- Rate limiting (future)

## License

MIT