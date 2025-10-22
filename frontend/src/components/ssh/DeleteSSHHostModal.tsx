import { Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../../i18n'
import type { SSHHost } from '../../types/remote'

interface DeleteSSHHostModalProps {
  isOpen: boolean
  host: SSHHost | null
  onClose: () => void
  onConfirm: () => void
  isLoading: boolean
}

const DeleteSSHHostModal: Component<DeleteSSHHostModalProps> = (props) => {
  const { t } = useI18n()

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box">
          <h3 class="font-bold text-lg">{t('ssh.confirm_delete')}</h3>
          <p class="py-4">
            {t('ssh.delete_warning', { name: props.host?.name || '' })}
          </p>
          <div class="modal-action">
            <button 
              class="btn"
              onClick={props.onClose}
            >
              {t('common.cancel')}
            </button>
            <button 
              class="btn btn-error"
              onClick={props.onConfirm}
              disabled={props.isLoading}
            >
              <Show when={props.isLoading}>
                <span class="loading loading-spinner loading-sm"></span>
              </Show>
              {t('common.delete')}
            </button>
          </div>
        </div>
      </div>
    </Show>
  )
}

export default DeleteSSHHostModal
