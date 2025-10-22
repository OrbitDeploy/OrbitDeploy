import { createSignal, createMemo, Show } from 'solid-js'
import type { Component } from 'solid-js'
import type { GitHubAppManifestRequest, GitHubAppManifestResponse } from '../types/providerAuth'
import { getGithubAppManifestEndpoint } from '../api/endpoints/providerAuths'
import { apiGet } from '../api/apiClient'

interface GitHubAppIntegrationModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
  userId: number
}

const GitHubAppIntegrationModal: Component<GitHubAppIntegrationModalProps> = (props) => {
  const [manifestResponse, setManifestResponse] = createSignal<GitHubAppManifestResponse | null>(null)
  const [isLoading, setIsLoading] = createSignal(false)

  const serverUrl = createMemo(() => {
    if (typeof window !== 'undefined') {
      const port = window.location.port === '3000' ? '8285' : window.location.port
      return `${window.location.protocol}//${window.location.hostname}:${port}`
    }
    return 'http://localhost:8285'
  })

  const callbackUrl = createMemo(() => {
    if (typeof window !== 'undefined') {
      return `${window.location.protocol}//${window.location.host}/provider-auths?github_app_created=true`
    }
    return 'http://localhost:3000/provider-auths?github_app_created=true'
  })

  const handleGenerateAndJump = async () => {
    console.log('Button clicked: Starting manifest generation')
    console.log('Server URL:', serverUrl(), 'Callback URL:', callbackUrl())
    setIsLoading(true)

    const variables: GitHubAppManifestRequest = {
      serverUrl: serverUrl(),
      callbackUrl: callbackUrl()
    }

    const endpoint = getGithubAppManifestEndpoint()
    const params = new URLSearchParams({
      serverUrl: variables.serverUrl,
      callbackUrl: variables.callbackUrl
    })
    const url = `${endpoint.url}?${params.toString()}`

    try {
      const data = await apiGet<GitHubAppManifestResponse>(url)

      console.log('Manifest generated successfully:', data)
      setManifestResponse(data)

      const form = document.createElement('form')
      form.method = 'POST'
      form.action = `https://github.com/settings/apps/new?state=${encodeURIComponent(data.manifestUrl.split('state=')[1])}`
      form.target = '_blank'
      const manifestInput = document.createElement('input')
      manifestInput.type = 'hidden'
      manifestInput.name = 'manifest'
      manifestInput.value = JSON.stringify(data.manifest)
      form.appendChild(manifestInput)
      document.body.appendChild(form)
      form.submit()
      document.body.removeChild(form)

      props.onSuccess('清单生成成功，已跳转到 GitHub 创建应用！')
    } catch (err: any) {
      console.error('Manifest generation failed:', err)
      props.onError(`生成 GitHub App 清单失败: ${err.message}`)
    } finally {
      setIsLoading(false)
    }
  }

  const handleClose = () => {
    setManifestResponse(null)
    props.onClose()
  }

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box w-11/12 max-w-3xl">
          <h3 class="font-bold text-lg mb-4">GitHub Apps 集成</h3>
          
          <div class="space-y-4">
            <div class="alert alert-info">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>点击下方按钮生成清单并跳转到 GitHub。在 GitHub 上，您可以编辑应用名称（已预填），然后安装应用。</span>
            </div>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text">服务器 URL</span>
              </label>
              <input
                type="text"
                class="input input-bordered"
                value={serverUrl()}
                disabled
              />
            </div>

            <div class="form-control">
              <label class="label">
                <span class="label-text">回调 URL</span>
              </label>
              <input
                type="text"
                class="input input-bordered"
                value={callbackUrl()}
                disabled
              />
            </div>

            {/* Show generated app name if available */}
            <Show when={manifestResponse()}>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">生成的 GitHub App 名称</span>
                </label>
                <input
                  type="text"
                  class="input input-bordered"
                  value={manifestResponse()?.manifest.name || ''}
                  disabled
                />
                <label class="label">
                  <span class="label-text-alt">可在 GitHub 上编辑</span>
                </label>
              </div>
            </Show>
          </div>

          <div class="modal-action">
            <button class="btn btn-ghost" onClick={handleClose}>
              取消
            </button>
            <button
              class="btn btn-primary"
              onClick={handleGenerateAndJump}
              disabled={isLoading()}
            >
              <Show when={isLoading()}>
                <span class="loading loading-spinner loading-sm mr-2"></span>
              </Show>
              <svg class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
              </svg>
              生成清单并跳转到 GitHub
            </button>
          </div>
        </div>
      </div>
    </Show>
  )
}

export default GitHubAppIntegrationModal