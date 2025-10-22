import { createSignal, Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import { useApiQuery, useApiMutation } from '../lib/apiHooks'
import { getCliApiUrl } from '../api/config'

// 设备上下文信息接口
interface DeviceContext {
  session_id: string
  os: string
  device_name: string
  public_ip: string
  request_timestamp: number
}

const CLIDeviceAuthPage: Component = () => {
  const { t } = useI18n()
  
  const [success, setSuccess] = createSignal(false)
  const [error, setError] = createSignal('')

  const urlParams = new URLSearchParams(window.location.search)
  const sessionId = () => urlParams.get('session_id')

  // 使用 useApiQuery 获取设备上下文信息
  const deviceQuery = useApiQuery<DeviceContext>(
    ['deviceContext', sessionId()],
    () => getCliApiUrl({ type: 'sessions', sessionId: sessionId()! }),
    {
      enabled: !!sessionId(),
    }
  )

  const deviceContext = () => deviceQuery.data
  const loading = () => deviceQuery.isPending


  const authorizeMutation = useApiMutation<unknown, { session_id: string; approved: boolean }>(
    getCliApiUrl('confirm'),
    {
      method: 'POST',
      onSuccess: () => {
        setSuccess(true)
        setError('')
        setTimeout(() => window.close(), 3000)
      },
      onError: (err: Error) => setError(err.message),
    }
  )

  const handleConfirm = (approved: boolean) => {
    const context = deviceContext()
    if (!context) return

    setError('')
    authorizeMutation.mutate({ 
      session_id: context.session_id, 
      approved 
    })
  }

  // 格式化时间戳
  const formatTimestamp = (timestamp: number) => {
    const date = new Date(timestamp * 1000)
    return date.toLocaleString(navigator.language, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZoneName: 'short'
    })
  }

  // 格式化操作系统显示
  const formatOS = (os: string) => {
    const osMap: Record<string, string> = {
      'windows': '💻 Windows',
      'linux': '🐧 Linux',
      'darwin': '🍎 macOS',
      'freebsd': '👹 FreeBSD'
    }
    return osMap[os.toLowerCase()] || `💻 ${os}`
  }

  return (
    <div class="min-h-screen bg-base-200 flex items-center justify-center p-4">
      <div class="card w-full max-w-lg bg-base-100 shadow-xl">
        <div class="card-body">
          <Show when={loading()}>
            <div class="text-center space-y-4">
              <span class="loading loading-spinner loading-lg"></span>
              <p>{t('common.loading')}</p>
            </div>
          </Show>

          <Show when={!loading() && (deviceQuery.error?.message || error())}>
            <div class="text-center space-y-4">
              <div class="text-error text-6xl">❌</div>
              <h3 class="text-lg font-semibold text-error">{t('common.error') || 'Error'}</h3>
              <p class="text-base-content/70">{deviceQuery.error?.message || error()}</p>
              <button
                class="btn btn-ghost"
                onClick={() => window.close()}
              >
                {t('cli_device_auth.cancel_close')}
              </button>
            </div>
          </Show>

          <Show when={!loading() && !(deviceQuery.error?.message || error()) && deviceContext() && !success()}>
            <div class="space-y-6">
              <div class="text-center">
                <h2 class="text-2xl font-bold mb-2">{t('cli_device_auth.title')}</h2>
                <p class="text-lg text-primary mb-4">{t('cli_device_auth.subtitle')}</p>
                <p class="text-base-content/70">{t('cli_device_auth.confirm_info')}</p>
              </div>

              <div class="bg-base-200 p-6 rounded-lg space-y-4">
                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.device_system')}:</span>
                  <span class="text-lg">{formatOS(deviceContext()?.os || '')}</span>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.device_name')}:</span>
                  <span class="font-mono bg-base-300 px-2 py-1 rounded">🆔 {deviceContext()?.device_name}</span>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.ip_address')}:</span>
                  <div class="text-right">
                    <div>🌍 {deviceContext()?.public_ip}</div>
                  </div>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.request_time')}:</span>
                  <span class="text-sm">🕒 {formatTimestamp(deviceContext()?.request_timestamp || 0)}</span>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.authorize_app')}:</span>
                  <span class="font-semibold text-primary">🔒 {t('cli_device_auth.app_name')}</span>
                </div>
              </div>

              <Show when={error()}>
                <div class="alert alert-error">
                  <span>{error()}</span>
                </div>
              </Show>

              <div class="flex gap-3">
                <button
                  class="btn btn-error flex-1"
                  disabled={authorizeMutation.isPending}
                  onClick={() => handleConfirm(false)}
                >
                  {t('cli_device_auth.deny_button')}
                </button>
                <button
                  class="btn btn-success flex-1"
                  disabled={authorizeMutation.isPending}
                  onClick={() => handleConfirm(true)}
                >
                  {authorizeMutation.isPending && <span class="loading loading-spinner loading-sm"></span>}
                  {authorizeMutation.isPending ? t('cli_device_auth.authorizing') : t('cli_device_auth.confirm_button')}
                </button>
              </div>
            </div>
          </Show>

          <Show when={success()}>
            <div class="text-center space-y-4">
              <div class="text-success text-6xl">✅</div>
              <h3 class="text-lg font-semibold text-success">{t('cli_device_auth.success_title')}</h3>
              <p class="text-base-content/70">
                {t('cli_device_auth.success_message')}
              </p>
              <p class="text-sm text-base-content/50">
                {t('cli_device_auth.auto_close_message')}
              </p>
            </div>
          </Show>

          <Show when={!success() && !loading() && !(deviceQuery.error?.message || error())}>
            <div class="divider">或</div>
            
            <div class="text-center">
              <button
                class="btn btn-ghost btn-sm"
                onClick={() => window.close()}
              >
                {t('cli_device_auth.cancel_close')}
              </button>
            </div>
          </Show>
        </div>
      </div>
    </div>
  )
}

export default CLIDeviceAuthPage