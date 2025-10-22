// Define the type for ProviderAuth
export interface ProviderAuth {
  uid: string
  platform: string
  applicationId?: number
  clientId: string
  redirectUri: string
  username: string | null
  appId: string | null
  slug: string | null
  webhookSecret: string | null
  installationId: number
  scopes: string | null
  isActive: boolean
  createdAt: string
  updatedAt: string
}
export interface CreateProviderAuthRequest {
  userId: number
  platform: 'github' | 'gitlab' | 'bitbucket' | 'gitea'
  applicationId?: number
  // GitHub platform specific fields (GitHub Apps only - OAuth Apps no longer supported)
  clientId: string  // For non-GitHub platforms
  clientSecret: string  // For non-GitHub platforms  
  redirectUri: string
  username: string  // For Bitbucket platform
  appPassword: string  // For Bitbucket platform
  appId: string  // For GitHub Apps (required for GitHub platform)
  slug: string   // For GitHub Apps
  privateKey: string  // For GitHub Apps (required for GitHub platform)
  webhookSecret: string
  installationId: number
  scopes: string
  isActive: boolean
}

export interface UpdateProviderAuthRequest {
  // GitHub platform specific fields (GitHub Apps only - OAuth Apps no longer supported)
  clientId: string  // For non-GitHub platforms
  clientSecret: string  // For non-GitHub platforms
  redirectUri: string
  username: string  // For Bitbucket platform
  appPassword: string  // For Bitbucket platform
  appId: string  // For GitHub Apps (required for GitHub platform)
  slug: string   // For GitHub Apps
  privateKey: string  // For GitHub Apps (required for GitHub platform)
  webhookSecret: string
  installationId: number
  scopes: string
  isActive: boolean
}

// GitHub Apps integration types
export interface GitHubAppManifestRequest {
  appName: string
  serverUrl: string
  callbackUrl: string
}

export interface GitHubAppManifest {
  name: string
  url: string
  hook_attributes: {
    url: string
  }
  public: boolean
  default_permissions: Record<string, string>
  default_events: string[]
  description: string
}

export interface GitHubAppManifestResponse {
  manifestUrl: string
  manifest: GitHubAppManifest
}

export interface GitHubAppCallbackRequest {
  code: string
  installationId?: number
  userId: number
}

export interface GitHubInstallRequest {
  installationId: number
}