# Project structure for Mafia Wars: Criminal Empire backend

```
mafia-wars-backend/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── app/                  # Application setup
│   │   └── app.go            # App initialization
│   ├── config/               # Configuration
│   │   └── config.go         # Configuration loading
│   ├── controller/           # API controllers
│   │   ├── auth.go           # Authentication controller
│   │   ├── market.go         # Market controller
│   │   ├── operations.go     # Operations controller
│   │   ├── player.go         # Player controller
│   │   └── territory.go      # Territory controller
│   ├── middleware/           # HTTP middlewares
│   │   ├── auth.go           # Authentication middleware
│   │   └── logging.go        # Logging middleware
│   ├── model/                # Database models
│   │   ├── market.go         # Market models
│   │   ├── operations.go     # Operations models
│   │   ├── player.go         # Player models
│   │   └── territory.go      # Territory models
│   ├── repository/           # Data access layer
│   │   ├── market.go         # Market repository
│   │   ├── operations.go     # Operations repository
│   │   ├── player.go         # Player repository
│   │   └── territory.go      # Territory repository
│   ├── service/              # Business logic
│   │   ├── auth.go           # Auth service
│   │   ├── market.go         # Market service
│   │   ├── operations.go     # Operations service
│   │   ├── player.go         # Player service
│   │   └── territory.go      # Territory service
│   └── util/                 # Utilities
│       ├── constants.go      # Constants
│       ├── jwt.go            # JWT helper
│       └── response.go       # Response formatter
├── migrations/               # Database migrations
│   └── 001_initial_schema.sql
├── pkg/                      # Reusable packages
│   ├── database/             # Database connection
│   │   └── postgres.go
│   └── logger/               # Logging
│       └── logger.go
├── configs/                  # Configuration files
│   ├── app.yaml              # Application config
│   └── mechanics.yaml        # Game mechanics config
├── scripts/                  # Helper scripts
│   ├── migrations.sh
│   └── seed.sh
├── go.mod                    # Go modules
└── go.sum                    # Module checksums
```