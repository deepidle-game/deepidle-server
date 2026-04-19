# DeepIdle Server - LLM Context

## Overview
Server-authoritative idle game backend in Go. Handles game simulation, authentication, inventory management, and resource gathering.

## Architecture
```
CLI Client → HTTP API → Go Server → MongoDB
                                ↓
                           State Response
```

## Key Components

### main.go
- Entry point using Fiber web framework
- Loads .env for configuration
- Connects to MongoDB
- Sets up routes and middleware

### Handlers (handlers/)
- **auth.go** - Login/signup with JWT tokens
- **character.go** - Character state, actions (cutting_wood, mining_rocks, fighting_monsters), resource claiming
- **inventory.go** - Inventory management, upgrade system
- **storage.go** - Global shared storage for all players

### Routes (routes/)
- routes.go - All API route registration

### Models (models/)
- Data structures for Character, Inventory, Storage

### Database (database/)
- MongoDB connection and seed data

### Middleware (middleware/)
- Auth middleware for protected routes

### State (state/)
- Server-side game state management

## API Endpoints

### Auth
- POST /api/auth/signup - Create account
- POST /api/auth/signin - Login

### Character
- GET /api/character/detail - Get character info
- POST /api/character/action - Set action (cutting_wood, mining_rocks, fighting_monsters)
- POST /api/character/claim - Claim accumulated resources
- PATCH /api/character/name - Update character name

### Inventory
- GET /api/inventory - Get inventory
- GET /api/inventory/upgrade-options - Get upgrade requirements
- POST /api/inventory/upgrade - Upgrade item

### Storage (Global, shared across all players)
- GET /api/storage - Get global storage
- POST /api/storage/deposit - Deposit item
- POST /api/storage/withdraw - Withdraw item

## Data Models

### Character
```json
{
  "id": " ObjectId",
  "username": "string",
  "name": "string",
  "level": 1,
  "current_action": "cutting_wood",
  "action_started_at": "timestamp",
  "last_claimed_at": "timestamp",
  "max_inventory_slots": 5
}
```

### InventoryItem
```json
{
  "item_id": "wooden_axe",
  "level": 1,
  "quantity": 0
}
```

### GlobalStorage
```json
{
  "item_id": "wood",
  "quantity": 100
}
```

## Game Logic

### Actions
- cutting_wood → produces wood (requires wooden_axe)
- mining_rocks → produces stone (requires wooden_pickaxe)
- fighting_monsters → produces gold (requires wooden_sword)

### Upgrade Formula
- wooden_axe Lv.N → N+1: wood = N×5, gold = N×10
- wooden_pickaxe Lv.N → N+1: stone = N×5, wood = N×5, gold = N×10
- wooden_sword Lv.N → N+1: wood = N×5, gold = N×10

### Inventory System
- 5 slots maximum
- Starts with: wooden_axe, wooden_pickaxe, wooden_sword (3 slots)
- Full inventory cannot claim resources

## Environment Variables
- PORT - Server port (default 3000)
- MONGO_URI - MongoDB connection string
- DB_NAME - Database name
- JWT_SECRET - JWT signing secret
- **CONTAINS CREDENTIALS - DO NOT COMMIT .env**

## State Machine (Job System)
```
IDLE → MOVING_TO_RESOURCE → GATHERING → INVENTORY_FULL → RETURNING_HOME → DEPOSITING → REPEAT
```

## Credentials
- MongoDB URI in .env
- JWT secret in .env
- **NEVER COMMIT .env FILE**