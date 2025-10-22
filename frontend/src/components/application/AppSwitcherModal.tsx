import type { Component } from 'solid-js'
import { useNavigate } from '@solidjs/router'
import type { Application } from '../../types/project'

interface AppSwitcherModalProps {
  isOpen: boolean
  currentAppUid: string
  projectUid: string
  projectName?: string  // Add optional project name
  applications: Application[]
  onClose: () => void
}

const AppSwitcherModal: Component<AppSwitcherModalProps> = (props) => {
  const navigate = useNavigate()

  const handleSwitchApp = (app: Application) => {
    if (app.uid !== props.currentAppUid) {
      // Use name-based route if project name is available, otherwise use UID
      if (props.projectName) {
        navigate(`/projects/${props.projectName}/apps/${app.name}`)
      } else {
        navigate(`/projects/${props.projectUid}/apps/${app.uid}`)
      }
    }
    props.onClose()
  }

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`}>
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">切换应用</h3>
        <p class="text-sm text-base-content/70 mb-4">
          选择要切换到的应用：
        </p>

        <div class="space-y-2 max-h-96 overflow-y-auto">
          {props.applications.map((app) => (
            <div
              class={`p-3 rounded-lg border cursor-pointer transition-colors ${
                app.uid === props.currentAppUid
                  ? 'border-primary bg-primary/10'
                  : 'border-base-300 hover:border-primary/50 hover:bg-base-200'
              }`}
              onClick={() => handleSwitchApp(app)}
            >
              <div class="flex items-center justify-between">
                <div class="flex-1">
                  <div class="font-medium">{app.name}</div>
                  <div class="text-sm text-base-content/70">
                    {app.description || '暂无描述'}
                  </div>
                  <div class="text-xs text-base-content/50 mt-1">
                    端口: {app.targetPort} | 状态: {app.status}
                  </div>
                </div>
                {app.uid === props.currentAppUid && (
                  <div class="text-primary">
                    <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        <div class="modal-action">
          <button class="btn" onClick={props.onClose}>
            取消
          </button>
        </div>
      </div>
    </div>
  )
}

export default AppSwitcherModal