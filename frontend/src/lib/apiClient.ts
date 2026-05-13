import axios from 'axios';

const configuredBaseUrl = (import.meta.env.VITE_API_BASE_URL || '').trim();
const isSecurePage = typeof window !== 'undefined' && window.location.protocol === 'https:';

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
