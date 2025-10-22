export interface RemoteContainer {
  id: string
  names: string[]
  image: string
  command: string
  status: string
  ports: string
  created_at: string
  host_id: number
  host_name: string
  host_address: string
}

export interface PodmanConnection {
  name: string
  uri: string
  identity: string
  default: boolean
}


export interface SSHHost {
  uid: string
  name: string
  addr: string
  port: number
  user: string
  description: string
  status: string
  region: string
  cpuCores: number
  memoryGB: number
  diskGB: number
  isActive: boolean
  createdAt: string
  updatedAt: string
  password?: string  // Only used in request, not in response
  private_key?: string  // Only used in request, not in response
}




export interface RemoteContainerManagementProps {
  hosts: SSHHost[]
  onRefreshHosts: () => void
}


export interface SSHHostRequest {
  name: string
  addr: string
  port: number
  user: string
  password: string
  private_key: string
  description: string
}
