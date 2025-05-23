# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mafia Wars: Criminal Empire is an asynchronous multiplayer web-based simulation game where players build and manage criminal organizations. The game features territory control, resource management, operations (missions), market trading, and campaign progression.

## Development Commands

### Backend (Go)
```bash
# Run backend server (development)
make run-be

# Run backend with air hot reload
make run-be-air

# Run database seeder
make run-be-seed

# Run tests
make test

# Run from backend directory
cd be/cmd/server && go run .
```

### Frontend (Vue.js)
```bash
# Run frontend development server
make run-fe

# Or from frontend directory
cd fe && npm run dev

# Build for production
cd fe && npm run build

# Type checking
cd fe && npm run type-check
```

### Full Development Environment
```bash
# Run both backend and frontend in new terminals
make run

# Using Docker Compose
docker-compose up -d
```

## Architecture Overview

### Backend Architecture (Go)
- **Framework**: Chi router with GORM ORM and PostgreSQL
- **Structure**: Clean architecture with repositories, services, and controllers
- **Configuration**: YAML-based config system with separate files for app, game mechanics, and territory data
- **Key Components**:
  - `internal/app/app.go`: Main application setup with dependency injection
  - `internal/config/`: Configuration loading with environment support
  - `internal/controller/`: HTTP handlers for API endpoints
  - `internal/service/`: Business logic layer with game mechanics
  - `internal/repository/`: Data access layer
  - `internal/model/`: Database models and entity definitions

### Frontend Architecture (Vue.js 3)
- **Framework**: Vue 3 with Composition API, TypeScript, and Pinia for state management
- **Routing**: Vue Router with authentication and region guards
- **Styling**: SCSS with centralized variables and mixins
- **Key Components**:
  - `src/stores/modules/`: Pinia stores for each game system
  - `src/services/`: API service layer using Axios
  - `src/components/`: Reusable Vue components
  - `src/views/`: Page-level components

### Game Systems Integration
The game features multiple interconnected systems:
- **Territory System**: Region-based map with hotspot control and income generation
- **Operations System**: Daily refreshing missions with resource costs and rewards
- **Market System**: Dynamic pricing for resource trading between players
- **Campaign System**: Structured story progression with chapters, missions, and POIs
- **Travel System**: Movement between regions with risk/reward mechanics
- **Authentication**: JWT-based with SSE (Server-Sent Events) for real-time updates

### Configuration System
Game mechanics are externalized in YAML files under `be/configs/`:
- `app.yaml`: Application and database configuration
- `game.yaml`: Game-specific settings and intervals
- `mechanics.yaml`: All game mechanics, formulas, and balance parameters
- `territory.yaml`: Hierarchical territory structure data
- `operations.yaml`: Operation definitions and parameters

### Database Design
Uses PostgreSQL with GORM auto-migration. Key entity relationships:
- Players have stats, resources, and territory ownership
- Territory hierarchy: Regions → Districts → Cities → Hotspots
- Operations and campaigns have completion tracking
- Market transactions and price history are logged

### API Design
RESTful API with consistent response wrapper format including payloads, errors, warnings, and game messages. All protected routes require JWT authentication with middleware.

### Real-time Features
Server-Sent Events (SSE) provide real-time updates for:
- Operation completions
- Territory changes
- Market price updates
- Notifications

## Development Notes

### Backend Development
- Game mechanics should be configured in YAML files, not hardcoded
- Use the established service provider pattern for extensibility
- All API responses use consistent wrapper format
- Database migrations are handled automatically via GORM

### Frontend Development
- Follow Vue 3 Composition API patterns with `<script setup>`
- Use centralized SCSS variables and mixins for consistent styling
- Implement responsive design for mobile compatibility
- Route guards handle authentication and region requirements automatically

### Testing
- Backend tests: `make test` or `go test ./...`
- Frontend type checking: `cd fe && npm run type-check`

### Database Setup
- Local development uses `scripts/setup-db.sh`
- Docker setup uses postgres service in docker-compose.yml
- Database name: `mwce` (local) or `mafia_wars` (Docker)

### Environment Configuration
- Backend runs on port 8000 (configurable in app.yaml)
- Frontend runs on port 3000
- Database on port 5432
- CORS configured for localhost:3000

### Game Design Context
The game follows a core loop: Complete Operations → Acquire Resources → Expand Territory → Gain Influence/Reduce Heat → Repeat. Understanding this flow is essential when working on any game system modifications.