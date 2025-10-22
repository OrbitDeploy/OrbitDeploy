import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import { useSystemStats } from '../lib/useSystemStats'

const SystemMonitorPage: Component = () => {
  const { t } = useI18n()
  const { stats, connection, formatBytes, formatUptime, getMemoryUsagePercent, getCpuUsagePercent } = useSystemStats()

  return (
    <div class="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-base-content">{t('system_monitor.title')}</h1>
        <p class="text-base-content/70 mt-2">{t('system_monitor.description')}</p>
        
        {/* Connection Status */}
        <div class="mt-4">
          <div class="flex items-center gap-2">
            <div class={`w-3 h-3 rounded-full ${
              connection().status === 'connected' ? 'bg-success' :
              connection().status === 'connecting' || connection().status === 'reconnecting' ? 'bg-warning' :
              'bg-error'
            }`}></div>
            <span class="text-sm font-medium">
              {t(`system_monitor.connection.${connection().status}`)}
            </span>
            {connection().lastUpdate != null && (
              <span class="text-xs text-base-content/60">
                ({new Date(connection().lastUpdate as number).toLocaleTimeString()})
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Stats Grid */}
      {stats() ? (
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Memory Usage Card */}
          <div class="card bg-base-100 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-primary mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-6 h-6 stroke-current">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                </svg>
                {t('system_monitor.memory.title')}
              </h2>
              
              <div class="space-y-4">
                {/* Memory Usage Bar */}
                <div>
                  <div class="flex justify-between text-sm mb-2">
                    <span>{t('system_monitor.memory.usage_percent')}</span>
                    <span class="font-bold">{getMemoryUsagePercent()}%</span>
                  </div>
                  <div class="w-full bg-base-300 rounded-full h-3">
                    <div 
                      class={`h-3 rounded-full transition-all duration-300 ${
                        getMemoryUsagePercent() > 90 ? 'bg-error' :
                        getMemoryUsagePercent() > 70 ? 'bg-warning' :
                        'bg-success'
                      }`}
                      style={`width: ${getMemoryUsagePercent()}%`}
                    ></div>
                  </div>
                </div>
                
                {/* Memory Details */}
                <div class="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div class="text-base-content/70">{t('system_monitor.memory.used')}</div>
                    <div class="font-bold text-primary">{formatBytes(stats()!.MemoryUsed)}</div>
                  </div>
                  <div>
                    <div class="text-base-content/70">{t('system_monitor.memory.total')}</div>
                    <div class="font-bold">{formatBytes(stats()!.MemoryTotal)}</div>
                  </div>
                  <div>
                    <div class="text-base-content/70">{t('system_monitor.memory.free')}</div>
                    <div class="font-bold text-success">{formatBytes(stats()!.MemoryTotal - stats()!.MemoryUsed)}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* CPU Usage Card */}
          <div class="card bg-base-100 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-secondary mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-6 h-6 stroke-current">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"></path>
                </svg>
                {t('system_monitor.cpu.title')}
              </h2>
              
              <div class="space-y-4">
                {/* CPU Usage Bar */}
                <div>
                  <div class="flex justify-between text-sm mb-2">
                    <span>{t('system_monitor.cpu.usage_percent')}</span>
                    <span class="font-bold">{getCpuUsagePercent()}%</span>
                  </div>
                  <div class="w-full bg-base-300 rounded-full h-3">
                    <div 
                      class={`h-3 rounded-full transition-all duration-300 ${
                        getCpuUsagePercent() > 90 ? 'bg-error' :
                        getCpuUsagePercent() > 70 ? 'bg-warning' :
                        'bg-secondary'
                      }`}
                      style={`width: ${getCpuUsagePercent()}%`}
                    ></div>
                  </div>
                </div>
                
                {/* CPU Details */}
                <div class="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <div class="text-base-content/70">{t('system_monitor.cpu.cores')}</div>
                    <div class="font-bold text-secondary">{stats()!.CpuCore} ({stats()!.CpuCoreLogic})</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* System Information Card */}
          <div class="card bg-base-100 shadow-lg lg:col-span-2">
            <div class="card-body">
              <h2 class="card-title text-accent mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-6 h-6 stroke-current">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
                {t('system_monitor.system_info.title')}
              </h2>
              
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 text-sm">
                <div>
                  <div class="text-base-content/70">{t('system_monitor.system_info.hostname')}</div>
                  <div class="font-bold">{stats()!.HostName}</div>
                </div>
                <div>
                  <div class="text-base-content/70">{t('system_monitor.system_info.os')}</div>
                  <div class="font-bold">{stats()!.OS}</div>
                </div>
                <div>
                  <div class="text-base-content/70">{t('system_monitor.system_info.platform')}</div>
                  <div class="font-bold">{stats()!.Platform}</div>
                </div>
                <div>
                  <div class="text-base-content/70">{t('system_monitor.system_info.kernel_arch')}</div>
                  <div class="font-bold">{stats()!.KernelArch}</div>
                </div>
                <div>
                  <div class="text-base-content/70">{t('system_monitor.system_info.uptime')}</div>
                  <div class="font-bold text-accent">{formatUptime(stats()!.Uptime)}</div>
                </div>
                <div>
                  <div class="text-base-content/70">Host ID</div>
                  <div class="font-bold font-mono text-xs">{stats()!.HostId.substring(0, 12)}...</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : (
        /* Loading State */
        <div class="flex items-center justify-center min-h-[400px]">
          <div class="text-center">
            <span class="loading loading-spinner loading-lg"></span>
            <p class="mt-4 text-base-content/70">{t('system_monitor.connection.connecting')}</p>
          </div>
        </div>
      )}
    </div>
  )
}

export default SystemMonitorPage