import { createSignal, onMount } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import { useApiMutation } from '../lib/apiHooks'
import { apiMutate } from '../lib/apiClient'

const CLIAuthorizePage: Component = () => {
  const { t } = useI18n()
  
  const [userCode, setUserCode] = createSignal('')
  const [loading, setLoading] = createSignal(false)
  const [success, setSuccess] = createSignal(false)
  const [error, setError] = createSignal('')

  // Get user code from URL parameters if provided
  onMount(() => {
    const urlParams = new URLSearchParams(window.location.search)
    const codeParam = urlParams.get('user_code')
    if (codeParam) {
      setUserCode(codeParam)
    }
  })

  const authorizeMutation = useApiMutation<unknown, { user_code: string }>(
    (payload) => apiMutate('/api/cli/authorize/device', { method: 'POST', body: payload }),
    {
      onSuccess: () => {
        setSuccess(true)
        setError('')
        setTimeout(() => window.close(), 3000)
      },
      onError: (err) => setError(err.message),
    }
  )

  const handleAuthorize = () => {
    if (!userCode().trim()) {
      setError('请输入用户代码')
      return
    }

    /* loading handled by mutation */
    setError('')

    authorizeMutation.mutate({ user_code: userCode().trim() })
  }

  return (
    <div class="min-h-screen bg-base-200 flex items-center justify-center p-4">
      <div class="card w-full max-w-md bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-center justify-center mb-6">
            CLI 设备授权
          </h2>
          
          {!success() ? (
            <>
              <div class="space-y-4">
                <div class="text-center text-base-content/70">
                  <p class="mb-2">请输入 CLI 工具显示的用户代码来授权设备访问</p>
                </div>
                
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">用户代码</span>
                  </label>
                  <input
                    type="text"
                    placeholder="输入用户代码"
                    class="input input-bordered w-full text-center font-mono tracking-wider"
                    value={userCode()}
                    onInput={(e) => setUserCode(e.currentTarget.value)}
                    style="text-transform: uppercase;"
                  />
                </div>

                {error() && (
                  <div class="alert alert-error">
                    <span>{error()}</span>
                  </div>
                )}

                <div class="form-control mt-6">
                  <button
                    class="btn btn-primary"
                    disabled={authorizeMutation.isPending || !userCode().trim()}
                    onClick={handleAuthorize}
                  >
                    {loading() && <span class="loading loading-spinner loading-sm"></span>}
                    {loading() ? '授权中...' : '授权设备'}
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div class="text-center space-y-4">
              <div class="text-success text-6xl">✓</div>
              <h3 class="text-lg font-semibold text-success">授权成功！</h3>
              <p class="text-base-content/70">
                设备已成功授权，您现在可以关闭此页面。
              </p>
              <p class="text-sm text-base-content/50">
                页面将在 3 秒后自动关闭...
              </p>
            </div>
          )}

          <div class="divider">或</div>
          
          <div class="text-center">
            <button
              class="btn btn-ghost btn-sm"
              onClick={() => window.close()}
            >
              取消并关闭
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default CLIAuthorizePage