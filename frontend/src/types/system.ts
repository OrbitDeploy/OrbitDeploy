export interface DiskPartition {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
}

export interface SystemStats {
  CreateAt: number
  HostId: string
  HostName: string
  Uptime: number
  OS: string
  Platform: string
  KernelArch: string
  CpuCore: number
  CpuCoreLogic: number
  CpuPercent: number[]
  MemoryTotal: number
  MemoryUsed: number
  PublicIpv4?: string
  PublicIpv6?: string
  disk_partitions?: DiskPartition[]
  disk_total?: number
  disk_used?: number
}

export interface ConnectionStatus {
  status: 'connecting' | 'connected' | 'disconnected' | 'reconnecting'
  lastUpdate?: number
}
