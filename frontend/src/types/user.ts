// User 相关类型定义
export interface User {
  id: number
  username: string
  created_at: string
  updated_at: string
}

// Example 相关类型定义
export interface Example {
  id: number
  name: string
  created_at: string
  updated_at: string
}

// Project Credential 相关类型定义
export interface ProjectCredential {
  id: number
  project_id: number
  credential_type: string
  credential_id: number
  is_default: boolean
  created_at: string
  updated_at: string
}