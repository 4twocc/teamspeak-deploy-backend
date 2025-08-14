export type TSInstance = {
  id: string
  name: string
  status: 'stopped' | 'starting' | 'running' | 'stopping' | 'error'
  host: string
  version: string
  created_at: string
  updated_at: string
  config: {
    server_name: string
    voice_port: number
    file_port: number
    query_port: number
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
