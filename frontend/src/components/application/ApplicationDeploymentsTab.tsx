import { Component, Show, For, createSignal } from 'solid-js'
import type { DeploymentHistory, Application } from '../../types/project'
import type { Deployment } from '../../types/deployment'
import { useApiQuery } from '../../api/apiHooksW.ts'
import { getApplicationByIdEndpoint, getApplicationDeploymentsEndpoint } from '../../api/endpoints'
import CreateDeploymentModal from './CreateDeploymentModal.tsx'
import DeploymentLogsModal from './DeploymentLogsModal.tsx'

interface ApplicationDeploymentsTabProps {
  applicationUid: string
}

const ApplicationDeploymentsTab: Component<ApplicationDeploymentsTabProps> = (props) => {
  const [showCreateModal, setShowCreateModal] = createSignal(false)
  const [showLogsModal, setShowLogsModal] = createSignal(false)
  const [selectedDeployment, setSelectedDeployment] = createSignal<Deployment | null>(null)

  // Fetch application data
  const appQuery = useApiQuery<Application>(
    () => ['applications', props.applicationUid],
    () => getApplicationByIdEndpoint(props.applicationUid).url,
    { enabled: () => !!props.applicationUid }
  )

  // Fetch deployment history
  const deploymentsQuery = useApiQuery<DeploymentHistory[]>(
    () => ['applications', props.applicationUid, 'deployments'],
    () => getApplicationDeploymentsEndpoint(props.applicationUid).url,
    { enabled: () => !!props.applicationUid }
  )

  // 打开日志模态框
  const openLogsModal = (deployment: DeploymentHistory) => {
    // 转换 DeploymentHistory 到 Deployment 类型
    const deploymentData: Deployment = {
      uid: deployment.uid,
      applicationUid: deployment.applicationUid,
      releaseUid: deployment.releaseUid,
      status: deployment.status,
      logText: deployment.logText,
      startedAt: deployment.startedAt,
      finishedAt: deployment.finishedAt,
      createdAt: deployment.createdAt,
      updatedAt: deployment.updatedAt,
      imageName: deployment.imageName,
      releaseStatus: deployment.releaseStatus
    }
    
    setSelectedDeployment(deploymentData)
    setShowLogsModal(true)
  }

  return (
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h3 class="text-lg font-semibold">部署历史</h3>
        <button 
          class="btn btn-primary btn-sm"
          onClick={() => setShowCreateModal(true)}
        >
          部署新版本
        </button>
      </div>

      <Show
        when={!deploymentsQuery.isPending}
        fallback={
          <div class="flex justify-center py-8">
            <span class="loading loading-spinner loading-lg"></span>
          </div>
        }
      >
        <Show
          when={deploymentsQuery.data && deploymentsQuery.data.length > 0}
          fallback={
            <div class="text-center py-8 text-base-content/70">
              <p>暂无部署记录</p>
            </div>
          }
        >
          <div class="overflow-x-auto">
            <table class="table table-zebra">
              <thead>
                <tr>
                  <th>部署ID</th>
                  <th>部署版本</th>
                  <th>主机端口</th>
                  <th>状态</th>
                  <th>开始时间</th>
                  <th>结束时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <For each={deploymentsQuery.data}>
                  {(deployment) => (
                    <tr>
                      <td>#{deployment.uid}</td>
                      <td class="font-mono text-sm">{deployment.imageName}</td>
                      <td>{deployment.systemPort || '-'}</td>
                      <td>
                        <span class={`badge ${
                          deployment.status === 'success' ? 'badge-success' :
                          deployment.status === 'failed' ? 'badge-error' :
                          deployment.status === 'running' ? 'badge-warning' :
                          'badge-info'
                        }`}>
                          {deployment.status}
                        </span>
                      </td>
                      <td>{new Date(deployment.startedAt).toLocaleString()}</td>
                      <td>{deployment.finishedAt ? new Date(deployment.finishedAt).toLocaleString() : '-'}</td>
                      <td>
                        <div class="flex gap-1">
                          <button 
                            class="btn btn-xs btn-outline"
                            onClick={() => openLogsModal(deployment)}
                          >
                            日志
                          </button>
                          <button class="btn btn-xs btn-outline">重启</button>
                        </div>
                      </td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </Show>
      </Show>

      {/* Create Deployment Modal */}
      <CreateDeploymentModal
        isOpen={showCreateModal()}
        onClose={() => setShowCreateModal(false)}
        application={appQuery.data || null}
      />

      {/* Deployment Logs Modal */}
      <DeploymentLogsModal
        isOpen={showLogsModal()}
        onClose={() => {
          setShowLogsModal(false)
          setSelectedDeployment(null)
        }}
        deployment={selectedDeployment()}
      />
    </div>
  )
}

export default ApplicationDeploymentsTab