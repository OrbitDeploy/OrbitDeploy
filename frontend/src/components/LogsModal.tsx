import type { Component } from 'solid-js'
import { useI18n } from '../i18n'

interface LogsModalProps {
  isOpen: boolean
  deploymentId: number | null
  logsText: string
  onClose: () => void
}

const LogsModal: Component<LogsModalProps> = (props) => {
  const { t } = useI18n()

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`} onClick={props.onClose}>
      <div class="modal-box max-w-5xl" onClick={(e) => e.stopPropagation()}>
        <h3 class="font-bold text-lg mb-4">
          {t('logs_modal.title')} #{props.deploymentId || ''}
        </h3>
        <div class="bg-base-200 p-4 rounded">
          <pre class="text-sm overflow-auto max-h-96 whitespace-pre-wrap">{props.logsText || t('logs_modal.loading')}</pre>
        </div>
        <div class="modal-action">
          <button class="btn" onClick={props.onClose}>
            {t('common.close')}
          </button>
        </div>
      </div>
    </div>
  )
}

export default LogsModal