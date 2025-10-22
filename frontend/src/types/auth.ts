// GitHub Token 相关类型定义
export interface GitHubToken {
  id: number
  name: string
  permissions: string
  expires_at?: string
  last_used_at?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

// Auth Token 相关类型定义
export interface AuthToken {
  id: number
  refresh_token_hash: string
  client_description: string
  expires_at: string
  created_at: string
  updated_at: string
}

// CLI Device Code 相关类型定义
export interface CLIDeviceCode {
  id: number
  device_code: string
  user_code: string
  is_authorized: boolean
  user_id?: number
  expires_at: string
  created_at: string
  updated_at: string
}