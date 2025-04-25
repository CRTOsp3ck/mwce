// src/services/api.ts

import axios, { AxiosResponse } from 'axios';

// Define GameMessage structure
export interface GameMessage {
  type: string;
  message: string;
}

// Define standard API response structure
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  gameMessage?: GameMessage;
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
  (response: AxiosResponse) => {
    // Log the API response for debugging
    // console.log("API Response", response.data);

    // Transform the response but preserve the AxiosResponse structure
    // This makes TypeScript happy while still giving components access to our parsed data
    // response.data = response.data as ApiResponse<any>;
    return response;
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

// Create typed request methods that handle the response unwrapping
// This is where we extract the ApiResponse from the AxiosResponse
const apiRequest = {
  get: async <T>(url: string, config = {}) => {
    const response = await api.get<any, AxiosResponse<ApiResponse<T>>>(url, config);
    return response.data;
  },

  post: async <T>(url: string, data = {}, config = {}) => {
    const response = await api.post<any, AxiosResponse<ApiResponse<T>>>(url, data, config);
    return response.data;
  },

  put: async <T>(url: string, data = {}, config = {}) => {
    const response = await api.put<any, AxiosResponse<ApiResponse<T>>>(url, data, config);
    return response.data;
  },

  delete: async <T>(url: string, config = {}) => {
    const response = await api.delete<any, AxiosResponse<ApiResponse<T>>>(url, config);
    return response.data;
  }
};

export default apiRequest;
