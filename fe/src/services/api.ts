// src/services/api.ts

import axios from 'axios';

// Create axios instance with default config
const api = axios.create({
  baseURL: 'http://localhost:8000/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
});

// Request interceptor for API calls
api.interceptors.request.use(
  (config) => {
    // Get the auth token from localStorage if it exists
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for API calls
api.interceptors.response.use(
  (response) => {
    // Returning the data from the axios response instead of the response itself
    return response.data;
  },
  async (error) => {
    const originalRequest = error.config;
    
    // Handle token refresh or redirect to login on auth errors
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      // For now, just redirect to login
      // In the future, we might implement token refresh here
      // localStorage.removeItem('auth_token');
      // window.location.href = '/login';
      
      return Promise.reject(error);
    }
    
    // Generic error handling
    if (error.response && error.response.data) {
      // Use the backend error message if available
      return Promise.reject(error.response.data);
    }
    
    return Promise.reject(error);
  }
);

export default api;