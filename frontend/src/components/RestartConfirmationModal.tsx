import { createSignal, Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'

interface RestartConfirmationModalProps {
  isOpen: boolean
  deploymentId: number | null
  serviceName?: string
  onClose: () => void
  onConfirm: (deploymentId: number) => Promise<void>
}

const RestartConfirmationModal: Component<RestartConfirmationModalProps> = (props) => {
  const { t } = useI18n()
  const [isConfirming, setIsConfirming] = createSignal(false)

  async function handleConfirm() {
    if (!props.deploymentId) return
    
    setIsConfirming(true)
    try {
      await props.onConfirm(props.deploymentId)
      props.onClose()
    } finally {
      setIsConfirming(false)
    }
  }

  function handleClose() {
    if (isConfirming()) return // Prevent closing while confirming
    props.onClose()
  }

  return (
    <div class={`modal ${props.isOpen && props.deploymentId ? 'modal-open' : ''}`}>
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4 text-warning">重启部署服务</h3>
        <div class="space-y-4">
          <div class="alert alert-warning">
            <div>
              <strong>确认操作：</strong> 即将重启部署服务
              <Show when={props.serviceName}>
                <div class="mt-2">
                  <strong>服务名称：</strong> {props.serviceName}
                </div>
              </Show>
              <ul class="list-disc list-inside mt-2 space-y-1">
                <li>服务将先停止后启动</li>
                <li>在重启期间服务将暂时不可用</li>
                <li>重启过程通常需要几秒钟时间</li>
              </ul>
            </div>
          </div>
          <p class="text-sm text-base-content/70">
            请确认是否要继续执行重启操作？
          </p>
        </div>
        <div class="modal-action">
          <button 
            class="btn btn-warning" 
            disabled={isConfirming()}
            onClick={() => void handleConfirm()}
          >
            <Show when={isConfirming()} fallback="确认重启">
              <span class="loading loading-spinner loading-sm"></span>
              重启中...
            </Show>
          </button>
          <button 
            class="btn" 
            disabled={isConfirming()}
            onClick={handleClose}
          >
            {t('common.cancel') || '取消'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default RestartConfirmationModal