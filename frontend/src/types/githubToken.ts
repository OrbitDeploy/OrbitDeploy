export interface GitHubToken {
  id: number
  name: string
  permissions?: string
  last_used_at?: string
  created_at: string
  expires_at?: string
  is_active: boolean
}
