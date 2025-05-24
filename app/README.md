# Mafia Wars: Criminal Empire

A web-based asynchronous multiplayer game where players build and manage criminal organizations in a shared city environment.

## Overview

*Mafia Wars: Criminal Empire* is a simulation game set in a stylized mid-20th century urban landscape. Players compete for influence, wealth, and power while navigating the complex web of alliances, betrayals, and law enforcement pressure.

## Features

- Build and manage your criminal empire
- Take over territory and defend your turf from rivals
- Complete operations to gain resources and reputation
- Buy and sell resources on the market
- Compete with other players for control of the city

## Technology Stack

### Backend
- Go (Golang)
- Chi router for HTTP routing
- GORM as the ORM
- PostgreSQL database

### Frontend
- Vue.js 3 with TypeScript
- Pinia for state management
- Vue Router
- Axios for API calls

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)
- PostgreSQL (for local development without Docker)

### Running with Docker Compose

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/mafia-wars.git
   cd mafia-wars
   ```

2. Start the application:
   ```bash
   docker-compose up -d
   ```

3. Access the application at http://localhost:3000

### Local Development Setup

#### Backend

1. Set up the database:
   ```bash
   cd mafia-wars-backend
   ./scripts/setup-db.sh
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Run the backend:
   ```bash
   go run cmd/server/main.go
   ```

#### Frontend

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Run the development server:
   ```bash
   npm run dev
   ```

4. Access the frontend at http://localhost:3000

## Project Structure

### Backend

```
mafia-wars-backend/
├── cmd/
│   ├── server/        # Application entry point
│   └── seed/          # Data seeding tools
├── configs/           # Configuration files
├── internal/          # Internal packages
│   ├── app/           # Application setup
│   ├── config/        # Configuration loading
│   ├── controller/    # API controllers
│   ├── middleware/    # HTTP middlewares
│   ├── model/         # Database models
│   ├── repository/    # Data access layer
│   ├── service/       # Business logic
│   └── util/          # Utilities
├── migrations/        # Database migrations
├── pkg/               # Reusable packages
│   ├── database/      # Database connection
│   └── logger/        # Logging
└── scripts/           # Helper scripts
```

### Frontend

```
frontend/
├── public/            # Static assets
├── src/
│   ├── assets/        # Application assets
│   ├── components/    # Vue components
│   ├── router/        # Vue Router configuration
│   ├── services/      # API services
│   ├── stores/        # Pinia stores
│   ├── types/         # TypeScript type definitions
│   └── views/         # Vue page components
├── .eslintrc.js       # ESLint configuration
└── vite.config.ts     # Vite configuration
```

## Game Mechanics

### Core Gameplay Loop

1. Complete Operations (PvE) → 
2. Acquire Resources → 
3. Expand Territory (PvP) → 
4. Gain Influence/Respect & Reduce Heat → 
5. Repeat

### Resources

- **Money**: Primary resource used for all activities
- **Crew Members**: Limited resource that enables parallel operations
- **Weapons**: Limited resource that enables parallel operations
- **Vehicles**: Limited resource that enables parallel operations

### Player Attributes

- **Respect**: Determined by success/failure rate of your territorial actions
- **Influence**: Determined by your overall territorial control
- **Title**: Determined by your overall respect & influence
- **Heat**: Negative attribute that accumulates through activities

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Game design inspired by classic territory control and mafia-themed games
- Built with appreciation for fans of strategy and simulation games