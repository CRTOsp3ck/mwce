-- Initial database schema for Mafia Wars: Criminal Empire

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Players table
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    title VARCHAR(50) NOT NULL,
    money INTEGER NOT NULL DEFAULT 0,
    crew INTEGER NOT NULL DEFAULT 0,
    max_crew INTEGER NOT NULL DEFAULT 25,
    weapons INTEGER NOT NULL DEFAULT 0,
    max_weapons INTEGER NOT NULL DEFAULT 30,
    vehicles INTEGER NOT NULL DEFAULT 0,
    max_vehicles INTEGER NOT NULL DEFAULT 12,
    respect INTEGER NOT NULL DEFAULT 0,
    influence INTEGER NOT NULL DEFAULT 0,
    heat INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_active TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Player statistics
CREATE TABLE IF NOT EXISTS player_stats (
    player_id UUID PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    total_operations_completed INTEGER NOT NULL DEFAULT 0,
    total_money_earned INTEGER NOT NULL DEFAULT 0,
    total_hotspots_controlled INTEGER NOT NULL DEFAULT 0,
    max_influence_achieved INTEGER NOT NULL DEFAULT 0,
    max_respect_achieved INTEGER NOT NULL DEFAULT 0,
    successful_takeovers INTEGER NOT NULL DEFAULT 0,
    failed_takeovers INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    read BOOLEAN NOT NULL DEFAULT FALSE
);

-- Achievements
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    criteria TEXT NOT NULL,
    reward TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Player achievements
CREATE TABLE IF NOT EXISTS player_achievements (
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (player_id, achievement_id)
);

-- Regions
CREATE TABLE IF NOT EXISTS regions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Districts
CREATE TABLE IF NOT EXISTS districts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Cities
CREATE TABLE IF NOT EXISTS cities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    district_id UUID NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Hotspots
CREATE TABLE IF NOT EXISTS hotspots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    city_id UUID NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    business_type VARCHAR(50) NOT NULL,
    is_legal BOOLEAN NOT NULL,
    controller_id UUID REFERENCES players(id) ON DELETE SET NULL,
    income INTEGER NOT NULL DEFAULT 0,
    pending_collection INTEGER NOT NULL DEFAULT 0,
    last_collection_time TIMESTAMP NOT NULL DEFAULT NOW(),
    crew INTEGER NOT NULL DEFAULT 0,
    weapons INTEGER NOT NULL DEFAULT 0,
    vehicles INTEGER NOT NULL DEFAULT 0,
    defense_strength INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Territory actions
CREATE TABLE IF NOT EXISTS territory_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    hotspot_id UUID NOT NULL REFERENCES hotspots(id) ON DELETE CASCADE,
    crew INTEGER NOT NULL DEFAULT 0,
    weapons INTEGER NOT NULL DEFAULT 0,
    vehicles INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN,
    money_gained INTEGER,
    money_lost INTEGER,
    crew_gained INTEGER,
    crew_lost INTEGER,
    weapons_gained INTEGER,
    weapons_lost INTEGER,
    vehicles_gained INTEGER,
    vehicles_lost INTEGER,
    respect_gained INTEGER,
    respect_lost INTEGER,
    influence_gained INTEGER,
    influence_lost INTEGER,
    heat_generated INTEGER,
    message TEXT,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Operations
CREATE TABLE IF NOT EXISTS operations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_special BOOLEAN NOT NULL DEFAULT FALSE,
    min_influence INTEGER NOT NULL DEFAULT 0,
    max_heat INTEGER NOT NULL DEFAULT 0,
    min_title VARCHAR(50),
    required_hotspot_types TEXT,
    crew INTEGER NOT NULL DEFAULT 0,
    weapons INTEGER NOT NULL DEFAULT 0,
    vehicles INTEGER NOT NULL DEFAULT 0,
    money INTEGER NOT NULL DEFAULT 0,
    reward_money INTEGER,
    reward_crew INTEGER,
    reward_weapons INTEGER,
    reward_vehicles INTEGER,
    reward_respect INTEGER,
    reward_influence INTEGER,
    reward_heat_reduction INTEGER,
    risk_crew_loss INTEGER,
    risk_weapons_loss INTEGER,
    risk_vehicles_loss INTEGER,
    risk_money_loss INTEGER,
    risk_heat_increase INTEGER,
    risk_respect_loss INTEGER,
    duration INTEGER NOT NULL,
    success_rate INTEGER NOT NULL,
    available_until TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Operation attempts
CREATE TABLE IF NOT EXISTS operation_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation_id UUID NOT NULL REFERENCES operations(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    crew INTEGER NOT NULL DEFAULT 0,
    weapons INTEGER NOT NULL DEFAULT 0,
    vehicles INTEGER NOT NULL DEFAULT 0,
    money INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN,
    money_gained INTEGER,
    money_lost INTEGER,
    crew_gained INTEGER,
    crew_lost INTEGER,
    weapons_gained INTEGER,
    weapons_lost INTEGER,
    vehicles_gained INTEGER,
    vehicles_lost INTEGER,
    respect_gained INTEGER,
    influence_gained INTEGER,
    heat_generated INTEGER,
    heat_reduced INTEGER,
    message TEXT,
    completion_time TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Market listings
CREATE TABLE IF NOT EXISTS market_listings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL UNIQUE,
    price INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    trend VARCHAR(50) NOT NULL,
    trend_percentage INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Market transactions
CREATE TABLE IF NOT EXISTS market_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    price INTEGER NOT NULL,
    total_cost INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    transaction_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Market price history
CREATE TABLE IF NOT EXISTS market_price_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resource_type VARCHAR(50) NOT NULL,
    price INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for faster queries
CREATE INDEX idx_notifications_player_id ON notifications(player_id);
CREATE INDEX idx_hotspots_controller_id ON hotspots(controller_id);
CREATE INDEX idx_hotspots_city_id ON hotspots(city_id);
CREATE INDEX idx_territory_actions_player_id ON territory_actions(player_id);
CREATE INDEX idx_territory_actions_hotspot_id ON territory_actions(hotspot_id);
CREATE INDEX idx_operation_attempts_player_id ON operation_attempts(player_id);
CREATE INDEX idx_operation_attempts_operation_id ON operation_attempts(operation_id);
CREATE INDEX idx_market_transactions_player_id ON market_transactions(player_id);
CREATE INDEX idx_market_price_history_resource_type ON market_price_history(resource_type);