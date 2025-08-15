// src/api/config.ts
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

export const API_ENDPOINTS = {
  // 实例管理
  INSTANCES: `${API_BASE_URL}/instances`,
  INSTANCE_DETAIL: (id: string) => `${API_BASE_URL}/instances/${id}`,
  INSTANCE_START: (id: string) => `${API_BASE_URL}/instances/${id}/start`,
  INSTANCE_STOP: (id: string) => `${API_BASE_URL}/instances/${id}/stop`,
  INSTANCE_RESTART: (id: string) => `${API_BASE_URL}/instances/${id}/restart`,
  INSTANCE_LOGS: (id: string) => `${API_BASE_URL}/instances/${id}/logs`,
} as const;

export default API_BASE_URL;
