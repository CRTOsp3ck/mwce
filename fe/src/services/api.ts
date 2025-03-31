// src/services/api.ts

import axios from 'axios';

// Define standard API response structure
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

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
    // Assuming backend returns { success: boolean, data: T, error?: {...} }
    const apiResponse = response.data as ApiResponse<any>;
    console.log(apiResponse)
    
    // If response is successful, return the data property
    // if (apiResponse.success && apiResponse.data !== undefined) {
    //   return apiResponse;
    // }
    
    // Otherwise, return the whole response
    return apiResponse;
  },
  async (error) => {
    const originalRequest = error.config;
    
    // Handle token refresh or redirect to login on auth errors
    if (error.response && error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      // For now, just redirect to login
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