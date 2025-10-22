import { createSignal, createMemo, Show } from 'solid-js'
import type { Component } from 'solid-js'
import type { ProviderAuth, CreateProviderAuthRequest } from '../types/providerAuth'

interface CreateEditProviderAuthModalProps {
  isOpen: boolean
  onClose: () => void
  editingAuth: ProviderAuth | null
  formData: Partial<CreateProviderAuthRequest>
  setFormData: (data: Partial<CreateProviderAuthRequest>) => void
  selectedPlatform: 'github' | 'gitlab' | 'bitbucket' | 'gitea'
  setSelectedPlatform: (platform: 'github' | 'gitlab' | 'bitbucket' | 'gitea') => void
  handleSubmit: () => void
  cancelEdit: () => void
  createMutation: { isLoading: boolean }
  updateMutation: { isLoading: boolean }
}

const CreateEditProviderAuthModal: Component<CreateEditProviderAuthModalProps> = (props) => {
  // Platform-related field display logic
  const showOAuthFields = createMemo(() => {
    const platform = props.selectedPlatform
    return platform === 'github' || platform === 'gitlab' || platform === 'gitea'
  })

  const showAppPasswordFields = createMemo(() => {
    return props.selectedPlatform === 'bitbucket'
  })

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box w-11/12 max-w-2xl">
          <h3 class="font-bold text-lg mb-4">
            {props.editingAuth ? '编辑授权' : '添加新授权'}
          </h3>

          <div class="space-y-4">
            {/* Platform selection */}
            <div class="form-control">
              <label class="label">
                <span class="label-text">平台类型</span>
              </label>
              <select
                class="select select-bordered"
                value={props.selectedPlatform}
                onChange={(e) => {
                  props.setSelectedPlatform(e.target.value as any)
                  props.setFormData({ ...props.formData, platform: e.target.value as any })
                }}
              >
                <option value="github">GitHub</option>
                <option value="gitlab">GitLab</option>
                <option value="bitbucket">Bitbucket</option>
                <option value="gitea">Gitea</option>
              </select>
            </div>

            {/* OAuth fields (GitHub, GitLab, Gitea) */}
            <Show when={showOAuthFields()}>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">Client ID</span>
                  <span class="label-text-alt text-error">*</span>
                </label>
                <input
                  type="text"
                  placeholder="输入OAuth应用的Client ID"
                  class="input input-bordered"
                  value={props.formData.clientId}
                  onInput={(e) => props.setFormData({ ...props.formData, clientId: e.target.value })}
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">Client Secret</span>
                  <span class="label-text-alt text-error">*</span>
                </label>
                <input
                  type="password"
                  placeholder="输入OAuth应用的Client Secret"
                  class="input input-bordered"
                  value={props.formData.clientSecret}
                  onInput={(e) => props.setFormData({ ...props.formData, clientSecret: e.target.value })}
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">Redirect URI</span>
                </label>
                <input
                  type="url"
                  placeholder="http://localhost:8285/api/providers/github/callback"
                  class="input input-bordered"
                  value={props.formData.redirectUri}
                  onInput={(e) => props.setFormData({ ...props.formData, redirectUri: e.target.value })}
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">Scopes</span>
                </label>
                <input
                  type="text"
                  placeholder="repo,read:user"
                  class="input input-bordered"
                  value={props.formData.scopes}
                  onInput={(e) => props.setFormData({ ...props.formData, scopes: e.target.value })}
                />
              </div>
            </Show>

            {/* App Password fields (Bitbucket) */}
            <Show when={showAppPasswordFields()}>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">用户名</span>
                  <span class="label-text-alt text-error">*</span>
                </label>
                <input
                  type="text"
                  placeholder="Bitbucket用户名"
                  class="input input-bordered"
                  value={props.formData.username}
                  onInput={(e) => props.setFormData({ ...props.formData, username: e.target.value })}
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">App Password</span>
                  <span class="label-text-alt text-error">*</span>
                </label>
                <input
                  type="password"
                  placeholder="Bitbucket应用密码"
                  class="input input-bordered"
                  value={props.formData.appPassword}
                  onInput={(e) => props.setFormData({ ...props.formData, appPassword: e.target.value })}
                />
              </div>
            </Show>

            {/* Is active toggle */}
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">启用授权</span>
                <input
                  type="checkbox"
                  class="toggle toggle-primary"
                  checked={props.formData.isActive}
                  onChange={(e) => props.setFormData({ ...props.formData, isActive: e.target.checked })}
                />
              </label>
            </div>
          </div>

          <div class="modal-action">
            <button class="btn btn-ghost" onClick={props.cancelEdit}>
              取消
            </button>
            <button
              class="btn btn-primary"
              onClick={props.handleSubmit}
              disabled={props.createMutation.isLoading || props.updateMutation.isLoading}
            >
              <Show when={props.createMutation.isLoading || props.updateMutation.isLoading}>
                <span class="loading loading-spinner loading-sm mr-2"></span>
              </Show>
              {props.editingAuth ? '更新' : '创建'}
            </button>
          </div>
        </div>
      </div>
    </Show>
  )
}

export default CreateEditProviderAuthModal
