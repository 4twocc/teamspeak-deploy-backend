export type InstanceStatus = 'stopped' | 'starting' | 'running' | 'stopping' | 'error';

export type TSInstance = {
  id: string
  name: string
  status: InstanceStatus
  host: string
  version: string
  created_at: string
  updated_at: string
  process_id: number
  config: {
    server_name: string
    welcome_msg: string
    max_clients: number
    default_server: number
    voice_port: number
    file_port: number
    query_port: number
    server_port: number
    query_admin_password: string
    server_admin_token: string
    log_queries: boolean
    log_client_cmds: boolean
  }
}

export interface ResourceUsage {
  cpu_percent: number;
  memory_mb: number;
  memory_percent: number;
  disk_usage_mb: number;
  network_in: number;
  network_out: number;
  timestamp: string;
}