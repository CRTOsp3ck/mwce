// src/services/authService.ts

import api from './api';

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
  territory: string;
}

export interface AuthResponse {
  token: string;
  player: {
    id: string;
    name: string;
    email: string;
  };
}

// Endpoints
const ENDPOINTS = {
  REGISTER: '/auth/register',
  LOGIN: '/auth/login'
};

export default {
  /**
   * Register a new user
   */
  register(data: RegisterRequest) {
    return api.post<AuthResponse>(ENDPOINTS.REGISTER, data);
  },
  
  /**
   * Login an existing user
   */
  login(data: LoginRequest) {
    return api.post<AuthResponse>(ENDPOINTS.LOGIN, data);
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
  }
};