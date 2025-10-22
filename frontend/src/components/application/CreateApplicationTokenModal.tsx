import { Component, createSignal, Show } from 'solid-js'
import { useI18n } from '../../i18n'
import { useApiMutation } from '../../api/apiHooksW.ts'
import { createApplicationTokenEndpoint } from '../../api/endpoints'

interface CreateApplicationTokenModalProps {
  isOpen: boolean
  onClose: () => void
   applicationUid: string
  applicationName: string
  onSuccess: (message: string) => void
  onError: (message: string) => void
}

interface CreateTokenRequest {
  name: string
  expires_at?: string
}

interface CreateTokenResponse {
  uid:  string
  applicationUid:  string
  name: string
  token: string // Only returned once
  expires_at?: string
  last_used_at?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

const CreateApplicationTokenModal: Component<CreateApplicationTokenModalProps> = (props) => {
  const { t } = useI18n()
  const [name, setName] = createSignal('')
  const [expiryDays, setExpiryDays] = createSignal(90)
  const [neverExpires, setNeverExpires] = createSignal(false)
  const [createdToken, setCreatedToken] = createSignal<string>('')
  const [showToken, setShowToken] = createSignal(false)

  // Create token mutation
  const createMutation = useApiMutation<CreateTokenResponse, CreateTokenRequest>(
    createApplicationTokenEndpoint(props.applicationUid),
    {
      onSuccess: (response) => {
        // The apiClient already extracts the data field, so response is the token data directly
        setCreatedToken(response.token)
        setShowToken(true)
        props.onSuccess(`Token "${response.name}" 创建成功`)
      },
      onError: (error) => {
        props.onError(error.message)
      }
    }
  )

  const handleSubmit = (e: Event) => {
    e.preventDefault()
    
    if (!name().trim()) {
      props.onError('请输入 Token 名称')
      return
    }

    let expires_at: string | undefined
    if (!neverExpires()) {
      const expiryDate = new Date()
      expiryDate.setDate(expiryDate.getDate() + expiryDays())
      expires_at = expiryDate.toISOString()
    }

    createMutation.mutate({
      name: name().trim(),
      expires_at
    })
  }

  const handleClose = () => {
    if (showToken()) {
      // If token was created, reset form and close
      setName('')
      setExpiryDays(90)
      setNeverExpires(false)
      setCreatedToken('')
      setShowToken(false)
    }
    props.onClose()
  }

  const copyToClipboard = () => {
    navigator.clipboard.writeText(createdToken())
      .then(() => props.onSuccess('Token 已复制到剪贴板'))
      .catch(() => props.onError('复制失败，请手动复制'))
  }

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box max-w-2xl">
          <h3 class="font-bold text-lg mb-4">
            为应用 "{props.applicationName}" 创建 Token
          </h3>

          <Show when={!showToken()}>
            <form onSubmit={handleSubmit} class="space-y-4">
              {/* Token Name */}
              <div class="form-control">
                <label class="label">
                  <span class="label-text">Token 名称 *</span>
                </label>
                <input
                  type="text"
                  class="input input-bordered w-full"
                  placeholder="例如：生产环境部署 Token"
                  value={name()}
                  onInput={(e) => setName(e.target.value)}
                  required
                />
                <label class="label">
                  <span class="label-text-alt">请使用描述性名称，便于管理和识别</span>
                </label>
              </div>

              {/* Expiry Settings */}
              <div class="form-control">
                <label class="label">
                  <span class="label-text">过期设置</span>
                </label>
                <div class="space-y-3">
                  <label class="cursor-pointer label justify-start">
                    <input
                      type="checkbox"
                      class="checkbox mr-3"
                      checked={neverExpires()}
                      onChange={(e) => setNeverExpires(e.target.checked)}
                    />
                    <span class="label-text">永不过期</span>
                  </label>

                  <Show when={!neverExpires()}>
                    <div class="flex items-center space-x-3">
                      <span class="text-sm">过期天数：</span>
                      <select
                        class="select select-bordered"
                        value={expiryDays()}
                        onChange={(e) => setExpiryDays(parseInt(e.target.value))}
                      >
                        <option value={7}>7 天</option>
                        <option value={30}>30 天</option>
                        <option value={90}>90 天</option>
                        <option value={180}>180 天</option>
                        <option value={365}>365 天</option>
                      </select>
                    </div>
                  </Show>
                </div>
              </div>

              {/* Submit Buttons */}
              <div class="modal-action">
                <button 
                  type="button" 
                  class="btn btn-ghost" 
                  onClick={handleClose}
                  disabled={createMutation.isPending}
                >
                  取消
                </button>
                <button 
                  type="submit" 
                  class="btn btn-primary"
                  disabled={createMutation.isPending}
                >
                  <Show when={createMutation.isPending}>
                    <span class="loading loading-spinner loading-sm mr-2"></span>
                  </Show>
                  创建 Token
                </button>
              </div>
            </form>
          </Show>

          {/* Token Display */}
          <Show when={showToken()}>
            <div class="space-y-4">
              <div class="alert alert-success">
                <div class="flex-1">
                  <h4 class="font-semibold">Token 创建成功！</h4>
                  <p class="text-sm mt-1">
                    请立即复制并保存此 Token，它只会显示一次。
                  </p>
                </div>
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text font-semibold">您的 Token</span>
                </label>
                <div class="flex">
                  <input
                    type="text"
                    class="input input-bordered flex-1 font-mono text-sm"
                    value={createdToken()}
                    readonly
                  />
                  <button
                    type="button"
                    class="btn btn-outline ml-2"
                    onClick={copyToClipboard}
                  >
                    复制
                  </button>
                </div>
              </div>

              <div class="bg-base-200 p-4 rounded-lg">
                <h5 class="font-semibold mb-2">使用方式</h5>
                <div class="space-y-2 text-sm">
                  <div>
                    <span class="font-medium">CLI 使用：</span>
                    <code class="block bg-base-300 p-2 rounded mt-1">
                      export ORBIT_TOKEN={createdToken()}<br/>
                      orbitctl deploy --app {props.applicationName}
                    </code>
                  </div>
                  <div>
                    <span class="font-medium">API 调用：</span>
                    <code class="block bg-base-300 p-2 rounded mt-1">
                      curl -H "Authorization: Bearer {createdToken()}" ...
                    </code>
                  </div>
                </div>
              </div>

              <div class="modal-action">
                <button 
                  type="button" 
                  class="btn btn-primary" 
                  onClick={handleClose}
                >
                  完成
                </button>
              </div>
            </div>
          </Show>
        </div>
      </div>
    </Show>
  )
}

export default CreateApplicationTokenModal