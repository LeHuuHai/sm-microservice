import axios from 'axios';

const api = axios.create({
  baseURL: 'http://192.168.9.250',
  headers: { 'Content-Type': 'application/json' },
});

export const tokenService = {
  getAccessToken: () => localStorage.getItem('access_token'),
  getRefreshToken: () => localStorage.getItem('refresh_token'),
  setTokens: (access: string, refresh: string) => {
    localStorage.setItem('access_token', access);
    localStorage.setItem('refresh_token', refresh);
  },
  clearTokens: () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  },
};

api.interceptors.request.use((config) => {
    const token = tokenService.getAccessToken();
    if (token && config.headers) config.headers['Authorization'] = `Bearer ${token}`;
    return config;
  }, (error) => Promise.reject(error)
);

api.interceptors.response.use((response) => response, async (error) => {
    const originalRequest = error.config;
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      const refreshToken = tokenService.getRefreshToken();
      if (refreshToken) {
        try {
          const res = await axios.post('http://192.168.9.250/auth/refresh', { refresh_token: refreshToken });
          if (res.data?.access_token) {
            tokenService.setTokens(res.data.access_token, res.data.refresh_token || refreshToken);
            originalRequest.headers['Authorization'] = `Bearer ${res.data.access_token}`;
            return api(originalRequest);
          }
        } catch (refreshError) {
          tokenService.clearTokens();
          window.location.href = '/login';
          return Promise.reject(refreshError);
        }
      } else {
        tokenService.clearTokens();
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default api;
