export interface Project {
  uid: string
  name: string
  gitRepository: string  // Updated to match API format
  isPrivate: boolean     // Updated to match API format
  sourceType?: string    // 仓库来源类型，默认 'github'
  description?: string
  createdAt?: string
  updatedAt?: string
  deletedAt?: string
  applications?: Application[]  // Added to match model relationship
}

// Application interface based on Go model
export interface Application {
  uid: string
  projectUid: string
  name: string
  description: string
  activeReleaseUid?: string
  repoUrl?: string
  buildDir: string
  buildType: string
  targetPort: number
  status: string
  volumes?: VolumeMount[]  // Changed from Record<string, any> to VolumeMount[]
  execCommand?: string
  autoUpdatePolicy?: string
  branch?: string
  createdAt?: string
  updatedAt?: string
}

export interface ApiListResponse<T> {
  success: boolean
  message?: string
  data?: T
}

// Application Log interface - Updated to match backend response
export interface ApplicationLog {
  deploymentUid?: string
  timestamp: string
  level: string
  source: string
  message: string
}

export interface ApplicationLogsResponse {
  logs: ApplicationLog[];
  hasMore: boolean;
  totalCount: number;
}


// Deployment History interface - Updated to match Go backend model
export interface DeploymentHistory {
  uid: string
  applicationUid: string
  releaseUid: string
  version?: string
  imageName: string
  systemPort: number
  status: string
  logText: string
  startedAt: string
  finishedAt?: string
  createdAt?: string
  updatedAt?: string
  releaseStatus?: string
  // New fields for running deployments overview
  domains?: string[]
  hostPort?: number
}

// VolumeMount interface for volume mounting
export interface VolumeMount {
  hostPath: string      // 主机路径（相对路径）
  containerPath: string // 容器内路径（绝对路径）
  readOnly?: boolean    // 是否只读挂载
}

// Configuration interface based on Go model
export interface Configuration {
  uid: string
  applicationUid: string
  version: number
  envVars: string
  isActive: boolean
  createdAt?: string
  updatedAt?: string
}

// Environment Variable interface
export interface EnvironmentVariable {
  uid: string
  configurationUid: string
  key: string
  value: string
  isEncrypted: boolean
  createdAt?: string
  updatedAt?: string
}

// Configuration with Environment Variables interface
export interface ConfigurationWithVariables {
  uid: string
  applicationUid: string
  version: number
  isActive: boolean
  createdAt?: string
  updatedAt?: string
  environmentVariables: EnvironmentVariable[]
}

// Request types for environment variables
export interface CreateEnvironmentVariableRequest {
  key: string
  value: string
  isEncrypted: boolean
}

export interface CreateConfigurationWithVariablesRequest {
  version: number
  isActive: boolean
  environmentVariables: CreateEnvironmentVariableRequest[]
}