// Routing 相关类型定义
export interface RoutingResponse {
  uid: string
  applicationUid: string
  domainName: string
  hostPort: number
  isActive: boolean
  createdAt: string
  updatedAt: string
}

// Routing 请求类型 (用于创建和更新)
export interface RoutingRequest {
  domainName: string
  hostPort: number
  isActive: boolean
}

// Add response wrapper for list API
export interface RoutingsResponse {
  data: {
    message: string
    routings: RoutingResponse[]
  }
  success: boolean
}