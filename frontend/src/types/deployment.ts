// Configuration 相关类型定义
export interface Configuration {
  uid: string
  applicationUid: string
  version: number
  envVars: string
  isActive: boolean
  createdAt?: string
  updatedAt?: string
}

// Deployment 相关类型定义
export interface Deployment {
  uid: string
  applicationUid: string
  releaseUid: string
  status: string
  logText: string
  startedAt: string
  finishedAt?: string
  createdAt?: string
  updatedAt?: string
  // Added fields from backend Release model
  version?: string  // Version from Release
  imageName?: string
  releaseStatus?: string
}

// DeploymentLog 结构化日志条目
export interface DeploymentLog {
  id: number
  timestamp: string
  level: string
  source: string
  message: string
}

// Release 相关类型定义
export interface Release {
  uid: string
  applicationUid: string
  version?: string  // Version field for release
  imageName: string
  buildSourceInfo: Record<string, any>  // JSONB type
  status: string
  systemPort?: number  // 系统分配的端口，可选字段
  createdAt?: string
  updatedAt?: string
}

// Environment 相关类型定义
export interface Environment {
  uid: string
  name: string
  description?: string
  created_at: string
  updated_at: string
}

// Deployment Spec 相关类型定义
export interface DeploymentSpecData {
  apiVersion: string
  kind: string
  metadata: {
    project: string
    environment: string
    name: string
  }
  spec: {
    strategy?: string
    replicas?: number
    containers: Container[]
    image?: {
      ref: string
    }
    env?: {
      setRef: string
    }
    secret?: {
      setRef: string
    }
  }
}

export interface Container {
  name: string
  publishPort?: number
  strategy?: string
  domains?: Array<{ host: string }>
  volumeMounts?: VolumeMount[]
}

export interface VolumeMount {
  name: string
  mountPath: string
  type?: string
  source?: string
}

export interface DeploymentRequest {
  environment: string
  spec: DeploymentSpecData
}

export interface DeploymentResponse {
  success: boolean
  message?: string
  data?: any
}