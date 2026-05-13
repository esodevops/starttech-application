import axios from 'axios';

const configuredBaseUrl = (import.meta.env.VITE_API_BASE_URL || '').trim();
const isSecurePage = typeof window !== 'undefined' && window.location.protocol === 'https:';
const AUTH_TOKEN_KEY = 'muchtodo_auth_token';

const API_BASE_URL = (() => {
  // On HTTPS pages, never use an insecure HTTP API URL (mixed content).
  if (configuredBaseUrl) {
    if (isSecurePage && configuredBaseUrl.startsWith('http://')) {
      return '';
    }
    return configuredBaseUrl;
  }

  // Dev keeps direct backend calls; production uses same-origin proxy.
  return import.meta.env.DEV ? 'http://localhost:8080' : '';
})();

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true, // Crucial for httpOnly cookies
});

export const setAuthToken = (token: string | null) => {
  if (typeof window === 'undefined') return;

  if (!token) {
    window.localStorage.removeItem(AUTH_TOKEN_KEY);
    return;
  }

  window.localStorage.setItem(AUTH_TOKEN_KEY, token);
};

export const getAuthToken = () => {
  if (typeof window === 'undefined') return null;
  return window.localStorage.getItem(AUTH_TOKEN_KEY);
};

apiClient.interceptors.request.use((config) => {
  const token = getAuthToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
