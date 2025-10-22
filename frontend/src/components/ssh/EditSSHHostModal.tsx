import { createSignal, createEffect, Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../../i18n'
import type { SSHHost, SSHHostRequest } from '../../types/remote'

interface EditSSHHostModalProps {
  isOpen: boolean
  host: SSHHost | null
  onClose: () => void
  onSubmit: (data: SSHHostRequest) => void
  isLoading: boolean
}

const EditSSHHostModal: Component<EditSSHHostModalProps> = (props) => {
  const { t } = useI18n()
  
  const [form, setForm] = createSignal<SSHHostRequest>({
    name: '',
    addr: '',
    port: 22,
    user: '',
    password: '',
    private_key: '',
    description: ''
  })

  createEffect(() => {
    if (props.host) {
      setForm({
        name: props.host.name,
        addr: props.host.addr,
        port: props.host.port,
        user: props.host.user,
        password: '', // Don't pre-fill sensitive data
        private_key: '', // Don't pre-fill sensitive data
        description: props.host.description
      })
    }
  })

  const resetForm = () => {
    setForm({
      name: '',
      addr: '',
      port: 22,
      user: '',
      password: '',
      private_key: '',
      description: ''
    })
  }

  const handleClose = () => {
    resetForm()
    props.onClose()
  }

  const handleSubmit = () => {
    props.onSubmit(form())
  }

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box max-w-2xl">
          <h3 class="font-bold text-lg mb-4">{t('ssh.edit_host')}</h3>
          
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('ssh.name')} *</span>
              </label>
              <input
                type="text"
                placeholder={t('ssh.name_placeholder')}
                class="input input-bordered"
                value={form().name}
                onInput={(e) => setForm({...form(), name: e.target.value})}
              />
            </div>

            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('ssh.address')} *</span>
              </label>
              <input
                type="text"
                placeholder={t('ssh.address_placeholder')}
                class="input input-bordered"
                value={form().addr}
                onInput={(e) => setForm({...form(), addr: e.target.value})}
              />
            </div>

            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('ssh.user')} *</span>
              </label>
              <input
                type="text"
                placeholder={t('ssh.user_placeholder')}
                class="input input-bordered"
                value={form().user}
                onInput={(e) => setForm({...form(), user: e.target.value})}
              />
            </div>

            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('ssh.port')}</span>
              </label>
              <input
                type="number"
                placeholder="22"
                class="input input-bordered"
                value={form().port}
                onInput={(e) => setForm({...form(), port: parseInt(e.target.value) || 22})}
              />
            </div>
          </div>

          <div class="form-control mt-4">
            <label class="label">
              <span class="label-text">{t('ssh.password')}</span>
            </label>
            <input
              type="password"
              placeholder={t('ssh.password_placeholder')}
              class="input input-bordered"
              value={form().password}
              onInput={(e) => setForm({...form(), password: e.target.value})}
            />
          </div>

          <div class="form-control mt-4">
            <label class="label">
              <span class="label-text">{t('ssh.private_key')}</span>
            </label>
            <textarea
              placeholder={t('ssh.private_key_placeholder')}
              class="textarea textarea-bordered h-32"
              value={form().private_key}
              onInput={(e) => setForm({...form(), private_key: e.target.value})}
            ></textarea>
          </div>

          <div class="form-control mt-4">
            <label class="label">
              <span class="label-text">{t('ssh.host_description')}</span>
            </label>
            <input
              type="text"
              placeholder={t('ssh.description_placeholder')}
              class="input input-bordered"
              value={form().description}
              onInput={(e) => setForm({...form(), description: e.target.value})}
            />
          </div>

          <div class="modal-action">
            <button 
              class="btn"
              onClick={handleClose}
            >
              {t('common.cancel')}
            </button>
            <button 
              class="btn btn-primary"
              onClick={handleSubmit}
              disabled={props.isLoading}
            >
              <Show when={props.isLoading}>
                <span class="loading loading-spinner loading-sm"></span>
              </Show>
              {t('common.save')}
            </button>
          </div>
        </div>
      </div>
    </Show>
  )
}

export default EditSSHHostModal
