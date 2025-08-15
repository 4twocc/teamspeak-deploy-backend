// src/api/instance.ts
import { API_ENDPOINTS } from './config';
import type { TSInstance } from '@/types/Instance';

export interface CreateInstanceInput {
  name: string;
  config: {
    server_name: string;
    welcome_msg?: string;
    max_clients?: number;
    voice_port?: number;
    file_port?: number;
    query_port?: number;
    server_port?: number;
    query_admin_password?: string;
    server_admin_token?: string;
  };
}

export interface InstanceListResponse {
  data: TSInstance[];
  meta: {
    page: number;
    page_size: number;
    total: number;
    pages: number;
  };
}

export interface InstanceResponse {
  data: TSInstance;
}

// 获取实例列表
export const getInstances = async (page = 1, pageSize = 10): Promise<InstanceListResponse> => {
  const response = await fetch(`${API_ENDPOINTS.INSTANCES}?page=${page}&page_size=${pageSize}`);
  if (!response.ok) {
    throw new Error('获取实例列表失败');
  }
  return response.json();
};

// 创建实例
export const createInstance = async (input: CreateInstanceInput): Promise<InstanceResponse> => {
  const response = await fetch(API_ENDPOINTS.INSTANCES, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input),
  });
  
  if (!response.ok) {
    throw new Error('创建实例失败');
  }
  
  return response.json();
};

// 获取实例详情
export const getInstance = async (id: string): Promise<InstanceResponse> => {
  const response = await fetch(API_ENDPOINTS.INSTANCE_DETAIL(id));
  if (!response.ok) {
    throw new Error('获取实例详情失败');
  }
  return response.json();
};

// 更新实例
export const updateInstance = async (id: string, input: Partial<CreateInstanceInput>): Promise<InstanceResponse> => {
  const response = await fetch(API_ENDPOINTS.INSTANCE_DETAIL(id), {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input),
  });
  
  if (!response.ok) {
    throw new Error('更新实例失败');
  }
  
  return response.json();
};

// 删除实例
export const deleteInstance = async (id: string): Promise<void> => {
  const response = await fetch(API_ENDPOINTS.INSTANCE_DETAIL(id), {
    method: 'DELETE',
  });
  
  if (!response.ok) {
    throw new Error('删除实例失败');
  }
};

// 启动实例
export const startInstance = async (id: string): Promise<InstanceResponse> => {
  const response = await fetch(API_ENDPOINTS.INSTANCE_START(id), {
    method: 'POST',
  });
  
  if (!response.ok) {
    throw new Error('启动实例失败');
  }
  
  return response.json();
};

// 停止实例
export const stopInstance = async (id: string): Promise<InstanceResponse> => {
  const response = await fetch(API_ENDPOINTS.INSTANCE_STOP(id), {
    method: 'POST',
  });
  
  if (!response.ok) {
    throw new Error('停止实例失败');
  }
  
  return response.json();
};

// 重启实例
export const restartInstance = async (id: string): Promise<InstanceResponse> => {
  const response = await fetch(API_ENDPOINTS.INSTANCE_RESTART(id), {
    method: 'POST',
  });
  
  if (!response.ok) {
    throw new Error('重启实例失败');
  }
  
  return response.json();
};

// 获取实例日志
export const getInstanceLogs = async (id: string): Promise<{ data: string }> => {
  const response = await fetch(API_ENDPOINTS.INSTANCE_LOGS(id));
  if (!response.ok) {
    throw new Error('获取实例日志失败');
  }
  return response.json();
};
