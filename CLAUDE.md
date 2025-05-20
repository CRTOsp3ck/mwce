# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview: Mafia Wars: Criminal Empire

Mafia Wars: Criminal Empire is an asynchronous multiplayer web-based simulation game where players build and manage criminal organizations in a shared city environment. The game features territory control, operations, a market system, and player progression mechanics.

## Development Commands

### Backend (Go)

```bash
# Run the backend server
cd be/cmd/server && go run .

# Run the backend with hot reload (requires air to be installed)
cd be/cmd/server && air

# Run the data seeder
cd be/cmd/seed && go run .

# Run tests
cd be && go test ./...
```

### Frontend (Vue.js)

```bash
# Start the development server
cd fe && npm run dev

# Type checking
cd fe && npm run type-check

# Build for production
cd fe && npm run build

# Preview the production build
cd fe && npm run preview
```

### Docker

```bash
# Run the entire application with Docker Compose
docker-compose up -d

# Stop the application
docker-compose down
```

### All-in-one Command (using Makefile)

```bash
# Run both frontend and backend in separate terminals
make run

# Clean build artifacts
make clean
```

## Architecture Overview

### Backend (Go)

The backend follows a layered architecture:

1. **Controllers (`be/internal/controller/`)**: Handle HTTP requests and responses
2. **Services (`be/internal/service/`)**: Implement business logic
3. **Repositories (`be/internal/repository/`)**: Handle data access
4. **Models (`be/internal/model/`)**: Define data structures

Key components:
- **Config (`be/internal/config/`)**: Loads configuration from YAML files
- **Middleware (`be/internal/middleware/`)**: HTTP middleware functions
- **Scheduler (`be/internal/scheduler/`)**: Background tasks and scheduled jobs

Game mechanics and configuration are stored in YAML files in `be/configs/` to facilitate easier game balancing.

### Frontend (Vue.js 3)

The frontend is built with Vue.js 3 using TypeScript:

1. **Views (`fe/src/views/`)**: Page components
2. **Components (`fe/src/components/`)**: Reusable UI components
3. **Stores (`fe/src/stores/`)**: Pinia state management
4. **Services (`fe/src/services/`)**: API services
5. **Types (`fe/src/types/`)**: TypeScript type definitions

### Data Flow

1. HTTP requests → Controllers → Services → Repositories → Database
2. Database → Repositories → Services → Controllers → HTTP responses

For real-time updates, the backend uses Server-Sent Events (SSE) to push updates to the frontend.

## Game Systems

1. **Territory Control**: Players compete to control hotspots in the city map (regions → districts → cities → hotspots)
2. **Operations**: PvE missions that players can complete for resources and rewards
3. **Market**: System for buying and selling resources with dynamic pricing
4. **Player Attributes**: Respect, Influence, Title, and Heat (negative attribute)
5. **Resources**: Money, Crew Members, Weapons, and Vehicles

## Current Development Focus

The team is currently working on implementing a Campaign Mode which adds narrative-driven gameplay through interconnected missions that utilize existing game mechanics. Campaign Mode features:

1. Narrative Integration: Story content contextualizing game mechanics
2. Campaign-Specific Operations: Operations that only appear during specific campaign missions
3. Hotspots-as-POIs: Custom interaction points for dialogue and other campaign activities
4. Branching Story Paths: Player choices affecting story progression and rewards

## File Organization

- `be/`: Backend code (Golang)
- `fe/`: Frontend code (Vue.js 3 with TypeScript)
- `configs/`: Configuration files for game mechanics and settings
- `docker-compose.yml`: Docker Compose configuration

## Additional Instructions

- Think step by step.
- Break down large task and ask questions when clarifying is needed. 
- Code should always be professional grade and the full implementation without placeholders unless stated otherwise.
- Level of accuracy of task execution is paramount.
- Create a todo list when working on complex tasks to track progress and remain on track