// src/services/authService.ts

import api, { ApiResponse } from './api';

// Define types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export interface AuthResponse {
  token: string;
  player: {
    id: string;
    name: string;
    email: string;
    title: string;
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
    createdAt: string;
    lastActive: string;
  };
}

// Endpoints
const ENDPOINTS = {
  REGISTER: '/auth/register',
  LOGIN: '/auth/login',
  VALIDATE: '/auth/validate'
};

export default {
  /**
   * Register a new user
   */
  register(data: RegisterRequest) {
    return api.post<ApiResponse<AuthResponse>>(ENDPOINTS.REGISTER, data);
  },
  
  /**
   * Login an existing user
   */
  login(data: LoginRequest) {
    return api.post<ApiResponse<AuthResponse>>(ENDPOINTS.LOGIN, data);
  },
  
  /**
   * Validate token
   */
  validate() {
    return api.get<ApiResponse<{message: string; user_id: string}>>(ENDPOINTS.VALIDATE);
  },
  
  /**
   * Logout current user
   */
  logout() {
    // Remove token from localStorage
    localStorage.removeItem('auth_token');
  },
  
  /**
   * Check if user is authenticated
   */
  isAuthenticated() {
    return !!localStorage.getItem('auth_token');
  },
  
  /**
   * Save token to localStorage
   */
  saveToken(token: string) {
    localStorage.setItem('auth_token', token);
  }
};