import { Component, Show, For } from 'solid-js'
import type { ApplicationLogsResponse,ApplicationLog } from '../../types/project'
import { useApiQuery } from '../../api/apiHooksW.ts'
import { getApplicationLogsEndpoint } from '../../api/endpoints'

interface ApplicationLogsTabProps {
  applicationUid: string
}

const ApplicationLogsTab: Component<ApplicationLogsTabProps> = (props) => {
  // Fetch logs data internally



const logsQuery = useApiQuery<ApplicationLogsResponse>(
    () => ['applications', props.applicationUid, 'logs'],
    () => getApplicationLogsEndpoint(props.applicationUid).url,
    { enabled: () => !!props.applicationUid }
  )
  // Add debug logging


  return (
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h3 class="text-lg font-semibold">应用日志</h3>
        <div class="flex gap-2">
          <select class="select select-sm select-bordered">
            <option value="">所有级别</option>
            <option value="info">Info</option>
            <option value="warn">Warning</option>
            <option value="error">Error</option>
            <option value="debug">Debug</option>
          </select>
          <button class="btn btn-sm btn-outline">清空</button>
          <button class="btn btn-sm btn-outline">下载</button>
        </div>
      </div>

      <Show
        when={!logsQuery.isPending}
        fallback={
          <div class="flex justify-center py-8">
            <span class="loading loading-spinner loading-lg"></span>
          </div>
        }
      >
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <Show
              when={logsQuery.data?.logs && logsQuery.data.logs.length > 0}
              fallback={
                <div class="text-center py-8 text-base-content/70">
                  <p>暂无日志记录</p>
                </div>
              }
            >
              <div class="space-y-2 max-h-96 overflow-y-auto">
                <For each={logsQuery.data?.logs}>
                  {(log) => (
                    <div class="flex gap-4 text-sm font-mono bg-base-200 p-2 rounded">
                      <span class="text-base-content/50 min-w-max">
                        {new Date(log.timestamp).toLocaleString()}
                      </span>
                      <span class={`min-w-max font-bold ${
                        log.level === 'error' ? 'text-error' :
                        log.level === 'warn' ? 'text-warning' :
                        log.level === 'info' ? 'text-info' :
                        'text-base-content/70'
                      }`}>
                        [{log.level.toUpperCase()}]
                      </span>
                      <span class="min-w-max text-base-content/70">[{log.source}]</span>
                      <span class="flex-1">{log.message}</span>
                    </div>
                  )}
                </For>
              </div>
            </Show>
          </div>
        </div>
      </Show>
    </div>
  )
}

export default ApplicationLogsTab