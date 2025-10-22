import { Component, createSignal } from 'solid-js'
import type { Application, DeploymentHistory } from '../../types/project'
import { useApiQuery } from '../../api/apiHooksW.ts'
import { getApplicationDeploymentsEndpoint } from '../../api/endpoints'
import CreateDeploymentModal from './CreateDeploymentModal.tsx'

interface ApplicationOverviewTabProps {
  applicationUid: string
  currentApp: Application | undefined
  appIdentifier?: string | number // Add this to pass name/ID from parent
}

const ApplicationOverviewTab: Component<ApplicationOverviewTabProps> = (props) => {
  const [showCreateModal, setShowCreateModal] = createSignal(false)

  // Fetch deployment history
  const deploymentsQuery = useApiQuery<DeploymentHistory[]>(
    () => ['applications', props.applicationUid, 'deployments'],
    () => getApplicationDeploymentsEndpoint(props.applicationUid).url,
    { enabled: () => !!props.applicationUid }
  )

  // Fetch running deployments for overview
  // const runningDeploymentsQuery = useApiQuery<DeploymentHistory[]>(
  //   () => ['applications', props.appIdentifier, 'runningDeployments'], // Use identifier for cache key
  //   () => getApplicationsApiUrl({ type: 'runningDeployments', identifier: props.appIdentifier }), // Pass identifier (name or ID)
  //   { 
  //     enabled: () => !!props.appIdentifier,
  //     refetchInterval: 5000 // Refresh every 5 seconds
  //   }
  // )

  const appDeployments = () => deploymentsQuery.data || []
  // const runningDeployments = () => runningDeploymentsQuery.data || []

  return (
    <div class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Basic Info Card */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <h3 class="card-title">基本信息</h3>
            <div class="space-y-2">
              <div class="flex justify-between">
                <span class="text-base-content/70">名称:</span>
                <span>{props.currentApp?.name}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-base-content/70">描述:</span>
                <span>{props.currentApp?.description || '-'}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-base-content/70">内部端口:</span>
                <span class="badge badge-outline">{props.currentApp?.targetPort}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-base-content/70">分支:</span>
                <span class="badge badge-outline">{props.currentApp?.branch}</span>
              </div>

              <div class="flex justify-between">
                <span class="text-base-content/70">状态:</span>
                <span class={`badge ${props.currentApp?.status === 'running' ? 'badge-success' :
                    props.currentApp?.status === 'stopped' ? 'badge-error' :
                      'badge-warning'
                  }`}>
                  {props.currentApp?.status}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Quick Actions Card */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <h3 class="card-title">快速操作</h3>
            <div class="space-y-2">
              <button class="btn btn-primary btn-sm w-full">重启应用</button>
              <button class="btn btn-outline btn-sm w-full">查看状态</button>
              <button
                class="btn btn-outline btn-sm w-full"
                onClick={() => setShowCreateModal(true)}
              >
                部署新版本
              </button>
            </div>
          </div>
        </div>

        {/* Stats Card */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <h3 class="card-title">统计信息</h3>
            <div class="space-y-2">
              <div class="flex justify-between">
                <span class="text-base-content/70">部署次数:</span>
                <span>{appDeployments().length}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-base-content/70">创建时间:</span>
                <span>{props.currentApp?.createdAt ? new Date(props.currentApp.createdAt).toLocaleDateString() : '-'}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-base-content/70">更新时间:</span>
                <span>{props.currentApp?.updatedAt ? new Date(props.currentApp.updatedAt).toLocaleDateString() : '-'}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Running Deployments Section */}
      {/* <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          <h3 class="card-title">当前运行的部署版本</h3>
          {runningDeployments().length === 0 ? (
            <div class="text-base-content/70 text-center py-4">
              暂无运行中的部署
            </div>
          ) : (
            <div class="overflow-x-auto">
              <table class="table table-zebra">
                <thead>
                  <tr>
                    <th>版本号</th>
                    <th>主机端口</th>
                    <th>部署时间</th>
                    <th>域名列表</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  {runningDeployments().map((deployment) => (
                    <tr>
                      <td>
                        <span class="badge badge-primary">{deployment.version}</span>
                      </td>
                      <td>
                        <span class="badge badge-outline">{deployment.hostPort || '-'}</span>
                      </td>
                      <td>
                        <div class="text-sm">
                          {deployment.startedAt 
                            ? new Date(deployment.startedAt).toLocaleString('zh-CN', {
                                year: 'numeric',
                                month: '2-digit',
                                day: '2-digit',
                                hour: '2-digit',
                                minute: '2-digit'
                              })
                            : '-'
                          }
                        </div>
                      </td>
                      <td>
                        <div class="flex flex-wrap gap-1">
                          {deployment.domains && deployment.domains.length > 0 ? (
                            deployment.domains.map((domain) => (
                              <span class="badge badge-ghost badge-sm">{domain}</span>
                            ))
                          ) : (
                            <span class="text-base-content/50">无域名</span>
                          )}
                        </div>
                      </td>
                      <td>
                        <span class="badge badge-success">{deployment.status}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div> */}

      {/* Create Deployment Modal */}
      <CreateDeploymentModal
        isOpen={showCreateModal()}
        onClose={() => setShowCreateModal(false)}
        application={props.currentApp || null}
      />
    </div>
  )
}

export default ApplicationOverviewTab