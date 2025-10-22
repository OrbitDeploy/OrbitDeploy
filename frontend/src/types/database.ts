// Database types
export type DatabaseType = 'postgresql' | 'mysql' | 'mongodb' | 'redis'

export type DatabaseStatus = 'pending' | 'running' | 'stopped' | 'failed'

export interface Database {
  uid: string
  name: string
  type: DatabaseType
  version: string
  custom_image?: string
  status: DatabaseStatus
  port: number
  internal_port: number
  username: string
  database_name: string
  data_path: string
  config_path: string
  is_remote: boolean
  ssh_host_uid?: string
  extra_config?: Record<string, unknown>
  last_check_at?: string
  created_at: string
  updated_at: string
}

export interface CreateDatabaseRequest {
  name: string
  type: DatabaseType
  version: string
  custom_image?: string
  port: number
  internal_port?: number
  username: string
  password: string
  database_name: string
  data_path: string
  config_path?: string
  is_remote?: boolean
  ssh_host_uid?: string
  extra_config?: Record<string, unknown>
}

export interface UpdateDatabaseRequest {
  port?: number
  username?: string
  password?: string
  data_path?: string
  config_path?: string
  extra_config?: Record<string, unknown>
}

export interface DatabaseConnectionInfo {
  host: string
  port: number
  user: string
  password?: string
  database: string
  connection_string: string
}
