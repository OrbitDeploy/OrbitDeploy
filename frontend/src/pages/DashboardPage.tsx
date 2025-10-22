import { createSignal, For } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import { useApiQuery } from '../lib/apiHooks'
import { getContainersApiUrl } from '../api/config'
import { useSystemStats } from '../lib/useSystemStats'
import { useRunningDeployments } from '../lib/useRunningDeployments'

const DashboardPage: Component = () => {
  const { t } = useI18n()

  // System stats from WebSocket
  const { stats: systemStats, connection, formatBytes, getMemoryUsagePercent, getCpuUsagePercent, getDiskUsagePercent, formatUptime } = useSystemStats()

  // Running deployments from SSE
  const { summary: runningDeploymentsSummary, connection: runningDeploymentsConnection, formatLastUpdated } = useRunningDeployments()

  // Mock dashboard data - in a real app, this would come from API
  const [stats] = createSignal({
    totalContainers: 0,
    todayViews: 0,
    totalUsers: 1,
    systemUptime: '0 hours'
  })

  // Fetch containers count for dashboard
  const containersQuery = useApiQuery<{ data: any[] }>(['containers'], getContainersApiUrl('list'))

  const currentStats = () => ({
    ...stats(),
    totalContainers: Array.isArray(containersQuery.data) ? containersQuery.data!.data.length : 0
  })

  return (
    <div class="p-6">
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-base-content">{t('dashboard.title')}</h1>
        <p class="text-base-content/70 mt-2">{t('dashboard.welcome')}</p>
      </div>

      {/* Stats Cards */}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div class="stat bg-base-100 shadow rounded-lg">
          <div class="stat-figure text-primary">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
            </svg>
          </div>
          <div class="stat-title">{t('dashboard.stats.total_containers')}</div>
          <div class="stat-value text-primary">{currentStats().totalContainers}</div>
          <div class="stat-desc">{t('dashboard.stats.total_containers_desc')}</div>
        </div>

        <div class="stat bg-base-100 shadow rounded-lg">
          <div class="stat-figure text-secondary">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4"></path>
            </svg>
          </div>
          <div class="stat-title">{t('dashboard.stats.today_views')}</div>
          <div class="stat-value text-secondary">{currentStats().todayViews}</div>
          <div class="stat-desc">{t('dashboard.stats.today_views_desc')}</div>
        </div>

        <div class="stat bg-base-100 shadow rounded-lg">
          <div class="stat-figure text-accent">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"></path>
            </svg>
          </div>
          <div class="stat-title">{t('dashboard.stats.admin_users')}</div>
          <div class="stat-value text-accent">{currentStats().totalUsers}</div>
          <div class="stat-desc">{t('dashboard.stats.admin_users_desc')}</div>
        </div>

        <div class="stat bg-base-100 shadow rounded-lg">
          <div class="stat-figure text-info">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
            </svg>
          </div>
          <div class="stat-title">{t('dashboard.stats.system_status')}</div>
          <div class="stat-value text-info">{t('dashboard.stats.system_status_value')}</div>
          <div class="stat-desc">{t('dashboard.stats.system_status_desc')}</div>
        </div>
      </div>
      <div class="stat bg-base-100 shadow rounded-lg   gap-6 mb-8">
        {/* <div class="stat bg-base-100 shadow rounded-lg grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8"> */}

        <div class="stat-figure text-success">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
        </div>
        <div class="stat-title">{t('dashboard.stats.running_deployments')}</div>
        <div class="stat-value text-success">{runningDeploymentsSummary()?.total_running || 0}</div>
        <div class="stat-desc">
          {runningDeploymentsSummary() && runningDeploymentsSummary()!.deployments.length > 0
            ? <ul class="list-disc list-inside">
              <For each={runningDeploymentsSummary()!.deployments} fallback={<li>No deployments</li>}>
                {(dep, index) => <li>{dep.app_name} ({dep.version})</li>}
              </For>
            </ul>
            : runningDeploymentsSummary() ? `${t('dashboard.stats.last_updated')}: ${formatLastUpdated(runningDeploymentsSummary()!.last_updated)}` : t('dashboard.stats.loading')}
        </div>
      </div>
      {/* System Monitoring Cards */}
      {systemStats() && (
        <div class="mb-8">
          <h2 class="text-xl font-bold text-base-content mb-4">{t('dashboard.system_monitor.title')}</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {/* Memory Usage Card */}
            <div class="card bg-base-100 shadow">
              <div class="card-body">
                <h3 class="card-title text-primary text-lg">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-5 h-5 stroke-current">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                  </svg>
                  {t('dashboard.system_monitor.memory')}
                </h3>
                <div class="flex justify-between items-center text-sm mb-2">
                  <span>{getMemoryUsagePercent()}%</span>
                  <span class="text-base-content/70">{formatBytes(systemStats()!.MemoryUsed)} / {formatBytes(systemStats()!.MemoryTotal)}</span>
                </div>
                <div class="w-full bg-base-300 rounded-full h-2">
                  <div
                    class={`h-2 rounded-full transition-all duration-300 ${getMemoryUsagePercent() > 90 ? 'bg-error' :
                        getMemoryUsagePercent() > 70 ? 'bg-warning' :
                          'bg-success'
                      }`}
                    style={`width: ${getMemoryUsagePercent()}%`}
                  ></div>
                </div>
              </div>
            </div>

            {/* CPU Usage Card */}
            <div class="card bg-base-100 shadow">
              <div class="card-body">
                <h3 class="card-title text-secondary text-lg">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-5 h-5 stroke-current">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"></path>
                  </svg>
                  {t('dashboard.system_monitor.cpu')}
                </h3>
                <div class="flex justify-between items-center text-sm mb-2">
                  <span>{getCpuUsagePercent()}%</span>
                  <span class="text-base-content/70">{systemStats()!.CpuCore} {t('dashboard.system_monitor.cores')}</span>
                </div>
                <div class="w-full bg-base-300 rounded-full h-2">
                  <div
                    class={`h-2 rounded-full transition-all duration-300 ${getCpuUsagePercent() > 90 ? 'bg-error' :
                        getCpuUsagePercent() > 70 ? 'bg-warning' :
                          'bg-secondary'
                      }`}
                    style={`width: ${getCpuUsagePercent()}%`}
                  ></div>
                </div>
              </div>
            </div>

            {/* Disk Usage Card */}
            <div class="card bg-base-100 shadow">
              <div class="card-body">
                <h3 class="card-title text-warning text-lg">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-5 h-5 stroke-current">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"></path>
                  </svg>
                  {t('dashboard.system_monitor.disk')}
                </h3>
                <div class="flex justify-between items-center text-sm mb-2">
                  <span>{getDiskUsagePercent()}%</span>
                  <span class="text-base-content/70">{formatBytes(systemStats()!.disk_used || 0)} / {formatBytes(systemStats()!.disk_total || 0)}</span>
                </div>
                <div class="w-full bg-base-300 rounded-full h-2">
                  <div
                    class={`h-2 rounded-full transition-all duration-300 ${getDiskUsagePercent() > 90 ? 'bg-error' :
                        getDiskUsagePercent() > 70 ? 'bg-warning' :
                          'bg-warning'
                      }`}
                    style={`width: ${getDiskUsagePercent()}%`}
                  ></div>
                </div>
              </div>
            </div>

            {/* System Info Card */}
            <div class="card bg-base-100 shadow">
              <div class="card-body">
                <h3 class="card-title text-accent text-lg">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-5 h-5 stroke-current">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                  </svg>
                  {t('dashboard.system_monitor.system_info')}
                </h3>
                <div class="space-y-1 text-sm">
                  <div class="flex justify-between">
                    <span class="text-base-content/70">{t('dashboard.system_monitor.uptime')}</span>
                    <span class="font-medium">{formatUptime(systemStats()!.Uptime)}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-base-content/70">{t('dashboard.system_monitor.os')}</span>
                    <span class="font-medium">{systemStats()!.OS}</span>
                  </div>
                </div>
                {/* Connection Status Indicator */}
                <div class="flex items-center gap-2 mt-2">
                  <div class={`w-2 h-2 rounded-full ${connection().status === 'connected' ? 'bg-success' :
                      connection().status === 'connecting' || connection().status === 'reconnecting' ? 'bg-warning' :
                        'bg-error'
                    }`}></div>
                  <span class="text-xs text-base-content/60">
                    {connection().status === 'connected' ? t('dashboard.system_monitor.connected') : t('dashboard.system_monitor.disconnected')}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Quick Actions */}
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title">{t('dashboard.quick_actions.title')}</h2>
            <div class="grid grid-cols-2 gap-4 mt-4">
              <button class="btn btn-primary btn-sm">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-4 h-4 stroke-current mr-2">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                </svg>
                {t('dashboard.quick_actions.add_example')}
              </button>
              <button class="btn btn-secondary btn-sm">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-4 h-4 stroke-current mr-2">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                </svg>
                {t('dashboard.quick_actions.system_settings')}
              </button>
            </div>
          </div>
        </div>

        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title">{t('dashboard.system_info.title')}</h2>
            <div class="mt-4">
              <div class="flex justify-between items-center py-2">
                <span class="text-sm opacity-70">{t('dashboard.system_info.version')}</span>
                <span class="text-sm">v{PACKAGE_VERSION}</span>
              </div>
              <div class="flex justify-between items-center py-2">
                <span class="text-sm opacity-70">{t('dashboard.system_info.database')}</span>
                <span class="text-sm">SQLite</span>
              </div>
              <div class="flex justify-between items-center py-2">
                <span class="text-sm opacity-70">{t('dashboard.system_info.uptime')}</span>
                <span class="text-sm">{currentStats().systemUptime}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default DashboardPage