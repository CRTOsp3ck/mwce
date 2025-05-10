// src/types/player.ts

export interface PlayerProfile {
  id: string;
  name: string;
  title: PlayerTitle;
  money: number;
  crew: number;
  maxCrew: number;
  weapons: number;
  maxWeapons: number;
  vehicles: number;
  maxVehicles: number;
  respect: number;
  influence: number;
  heat: number;
  controlledHotspots: number;
  totalHotspotCount: number;
  hourlyRevenue: number;
  pendingCollections: number;
  createdAt: string;
  lastActive: string;
}

export enum PlayerTitle {
  ASSOCIATE = "Associate",
  SOLDIER = "Soldier",
  CAPO = "Capo",
  UNDERBOSS = "Underboss",
  CONSIGLIERE = "Consigliere",
  BOSS = "Boss",
  GODFATHER = "Godfather",
}

export interface PlayerStats {
  totalOperationsCompleted: number;
  totalMoneyEarned: number;
  totalHotspotsControlled: number;
  maxInfluenceAchieved: number;
  maxRespectAchieved: number;
  successfulTakeovers: number;
  failedTakeovers: number;
}

export interface Notification {
  id: string;
  playerId: string;
  message: string;
  type: NotificationType;
  timestamp: string;
  read: boolean;
}

export enum NotificationType {
  TERRITORY = "territory",
  OPERATION = "operation",
  COLLECTION = "collection",
  HEAT = "heat",
  SYSTEM = "system",
  TRAVEL = "travel"
}
