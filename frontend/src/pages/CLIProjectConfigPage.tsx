import { createSignal, createEffect, onMount, Show, For } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import { useApiMutation } from '../lib/apiHooks'
import { apiMutate } from '../lib/apiClient'

// Types for project configuration
interface ProjectConfig {
  name: string
  description: string
  port: number
  domains: string[]
  env: Record<string, string>
  volumes: VolumeMount[]
  dockerfile?: string
  buildContext?: string
}

interface VolumeMount {
  hostPath: string
  containerPath: string
  readOnly: boolean
}

interface TomlPreview {
  app?: {
    name?: string
  }
  service?: {
    internalPort?: number
    publishPort?: number
  }
  env?: Record<string, string>
  volumes?: Array<{
    hostPath?: string
    containerPath?: string
    readOnly?: boolean
  }>
}

const CLIProjectConfigPage: Component = () => {
  const { t } = useI18n()
  
  // Page state
  const [sessionId, setSessionId] = createSignal('')
  const [expiresAt, setExpiresAt] = createSignal<Date | null>(null)
  const [timeLeft, setTimeLeft] = createSignal(0)
  const [expired, setExpired] = createSignal(false)
  const [submitted, setSubmitted] = createSignal(false)
  
  // Form state
  const [config, setConfig] = createSignal<ProjectConfig>({
    name: '',
    description: '',
    port: 8080,
    domains: [],
    env: {},
    volumes: []
  })
  const [tomlPreview, setTomlPreview] = createSignal<TomlPreview>({})
  const [loading, setLoading] = createSignal(false)
  const [error, setError] = createSignal('')
  
  // Environment variables input
  const [envKey, setEnvKey] = createSignal('')
  const [envValue, setEnvValue] = createSignal('')
  
  // Domains input
  const [newDomain, setNewDomain] = createSignal('')

  // Initialize page from URL parameters
  onMount(() => {
    const urlParams = new URLSearchParams(window.location.search)
    const sessionParam = urlParams.get('session')
    const tomlParam = urlParams.get('toml')
    const expiresParam = urlParams.get('expires')
    
    if (sessionParam) {
      setSessionId(sessionParam)
    }
    
    if (expiresParam) {
      const expiry = new Date(parseInt(expiresParam) * 1000)
      setExpiresAt(expiry)
    }
    
    if (tomlParam) {
      try {
        const decoded = atob(tomlParam)
        const parsed = parseTomlPreview(decoded)
        setTomlPreview(parsed)
        populateFromToml(parsed)
      } catch (err) {
        console.warn('Failed to parse TOML preview:', err)
      }
    }
    
    // Start countdown timer
    startCountdown()
  })

  // Countdown timer
  const startCountdown = () => {
    const timer = setInterval(() => {
      const expires = expiresAt()
      if (!expires) return
      
      const now = new Date()
      const diff = expires.getTime() - now.getTime()
      
      if (diff <= 0) {
        setExpired(true)
        setTimeLeft(0)
        clearInterval(timer)
      } else {
        setTimeLeft(Math.floor(diff / 1000))
      }
    }, 1000)
  }

  // Parse TOML preview (simplified parser for demo)
  const parseTomlPreview = (tomlText: string): TomlPreview => {
    const lines = tomlText.split('\n')
    const result: TomlPreview = {}
    
    lines.forEach(line => {
      line = line.trim()
      if (line.startsWith('name =')) {
        result.app = { name: line.split('=')[1].trim().replace(/['"]/g, '') }
      } else if (line.startsWith('internal_port =')) {
        result.service = { ...result.service, internalPort: parseInt(line.split('=')[1].trim()) }
      } else if (line.startsWith('http_port =')) {
        result.service = { ...result.service, publishPort: parseInt(line.split('=')[1].trim()) }
      }
    })
    
    return result
  }

  // Populate form from TOML preview
  const populateFromToml = (toml: TomlPreview) => {
    setConfig(prev => ({
      ...prev,
      name: toml.app?.name || prev.name,
      port: toml.service?.internalPort || toml.service?.publishPort || prev.port
    }))
  }

  // Format time remaining
  const formatTimeLeft = (seconds: number): string => {
    const minutes = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${minutes}:${secs.toString().padStart(2, '0')}`
  }

  // Add environment variable
  const addEnvVar = () => {
    if (envKey().trim() && envValue().trim()) {
      setConfig(prev => ({
        ...prev,
        env: { ...prev.env, [envKey().trim()]: envValue().trim() }
      }))
      setEnvKey('')
      setEnvValue('')
    }
  }

  // Remove environment variable
  const removeEnvVar = (key: string) => {
    setConfig(prev => {
      const newEnv = { ...prev.env }
      delete newEnv[key]
      return { ...prev, env: newEnv }
    })
  }

  // Add domain
  const addDomain = () => {
    if (newDomain().trim()) {
      setConfig(prev => ({
        ...prev,
        domains: [...prev.domains, newDomain().trim()]
      }))
      setNewDomain('')
    }
  }

  // Remove domain
  const removeDomain = (index: number) => {
    setConfig(prev => ({
      ...prev,
      domains: prev.domains.filter((_, i) => i !== index)
    }))
  }

  // Add volume mount
  const addVolumeMount = () => {
    setConfig(prev => ({
      ...prev,
      volumes: [...prev.volumes, { hostPath: '', containerPath: '', readOnly: false }]
    }))
  }

  // Remove volume mount
  const removeVolumeMount = (index: number) => {
    setConfig(prev => ({
      ...prev,
      volumes: prev.volumes.filter((_, i) => i !== index)
    }))
  }

  // Update volume mount
  const updateVolumeMount = (index: number, field: keyof VolumeMount, value: string | boolean) => {
    setConfig(prev => ({
      ...prev,
      volumes: prev.volumes.map((vol, i) => 
        i === index ? { ...vol, [field]: value } : vol
      )
    }))
  }

  // Add mutation for submitting configuration
  const submitMutation = useApiMutation<unknown, { session_id: string; config: ProjectConfig }>(
    (data) => apiMutate('/api/cli/configure/submit', { method: 'POST', body: data }),
    {
      onSuccess: () => {
        setSubmitted(true)
        setError('')
        // Auto-close page after successful submission
        setTimeout(() => {
          window.close()
        }, 3000)
      },
      onError: (err) => setError(err.message),
    }
  )

  // Submit configuration
  const handleSubmit = async () => {
    if (expired()) {
      setError(t('cli_project_config.error_session_expired'))
      return
    }

    if (!config().name.trim()) {
      setError(t('cli_project_config.error_project_name_required'))
      return
    }

    setLoading(true)
    setError('')

    // Use mutation instead of direct fetch
    submitMutation.mutate({
      session_id: sessionId(),
      config: config()
    })

    setLoading(false)  // Note: Loading is now handled by mutation.isPending
  }

  return (
    <div class="min-h-screen bg-base-200 p-4">
      <div class="max-w-4xl mx-auto">
        {/* Header with countdown */}
        <div class="card bg-base-100 shadow-xl mb-6">
          <div class="card-body">
            <div class="flex justify-between items-center">
              <h1 class="card-title text-2xl">{t('cli_project_config.title')}</h1>
              <div class={`badge ${expired() ? 'badge-error' : timeLeft() < 60 ? 'badge-warning' : 'badge-success'} badge-lg`}>
                {expired() ? t('cli_project_config.expired') : `${t('cli_project_config.time_left')}: ${formatTimeLeft(timeLeft())}`}
              </div>
            </div>
            <p class="text-base-content/70">
              {t('cli_project_config.description')}
            </p>
          </div>
        </div>

        {!submitted() && !expired() ? (
          <>
            {/* TOML Preview */}
            <Show when={Object.keys(tomlPreview()).length > 0}>
              <div class="card bg-base-100 shadow-xl mb-6">
                <div class="card-body">
                  <h2 class="card-title">{t('cli_project_config.toml_preview')}</h2>
                  <div class="bg-base-200 p-4 rounded-lg">
                    <pre class="text-sm">
                      {JSON.stringify(tomlPreview(), null, 2)}
                    </pre>
                  </div>
                </div>
              </div>
            </Show>

            {/* Configuration Form */}
            <div class="card bg-base-100 shadow-xl">
              <div class="card-body">
                <h2 class="card-title mb-4">{t('cli_project_config.project_config')}</h2>
                
                {error() && (
                  <div class="alert alert-error mb-4">
                    <span>{error()}</span>
                  </div>
                )}

                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  {/* Basic Information */}
                  <div class="space-y-4">
                    <h3 class="font-semibold text-lg">{t('cli_project_config.basic_info')}</h3>
                    
                    <div class="form-control">
                      <label class="label">
                        <span class="label-text">{t('cli_project_config.project_name_required')}</span>
                      </label>
                      <input
                        type="text"
                        class="input input-bordered"
                        value={config().name}
                        onInput={(e) => setConfig(prev => ({ ...prev, name: e.currentTarget.value }))}
                        placeholder={t('cli_project_config.project_name_placeholder')}
                      />
                    </div>

                    <div class="form-control">
                      <label class="label">
                        <span class="label-text">{t('cli_project_config.project_description')}</span>
                      </label>
                      <textarea
                        class="textarea textarea-bordered"
                        value={config().description}
                        onInput={(e) => setConfig(prev => ({ ...prev, description: e.currentTarget.value }))}
                        placeholder={t('cli_project_config.project_description_placeholder')}
                      />
                    </div>

                    <div class="form-control">
                      <label class="label">
                        <span class="label-text">{t('cli_project_config.port_number')}</span>
                      </label>
                      <input
                        type="number"
                        class="input input-bordered"
                        value={config().port}
                        onInput={(e) => setConfig(prev => ({ ...prev, port: parseInt(e.currentTarget.value) || 8080 }))}
                        min="1"
                        max="65535"
                      />
                    </div>
                  </div>

                  {/* Domains */}
                  <div class="space-y-4">
                    <h3 class="font-semibold text-lg">{t('cli_project_config.domain_config')}</h3>
                    
                    <div class="form-control">
                      <label class="label">
                        <span class="label-text">{t('cli_project_config.add_domain')}</span>
                      </label>
                      <div class="join">
                        <input
                          type="text"
                          class="input input-bordered join-item flex-1"
                          value={newDomain()}
                          onInput={(e) => setNewDomain(e.currentTarget.value)}
                          placeholder={t('cli_project_config.domain_placeholder')}
                        />
                        <button
                          type="button"
                          class="btn btn-primary join-item"
                          onClick={addDomain}
                          disabled={!newDomain().trim()}
                        >
                          {t('cli_project_config.add_button')}
                        </button>
                      </div>
                    </div>

                    <div class="space-y-2">
                      <For each={config().domains}>{(domain, index) => (
                        <div class="flex items-center gap-2">
                          <span class="flex-1 p-2 bg-base-200 rounded">{domain}</span>
                          <button
                            type="button"
                            class="btn btn-error btn-sm"
                            onClick={() => removeDomain(index())}
                          >
                            {t('cli_project_config.remove_button')}
                          </button>
                        </div>
                      )}</For>
                    </div>
                  </div>
                </div>

                {/* Environment Variables */}
                <div class="mt-6">
                  <h3 class="font-semibold text-lg mb-4">{t('cli_project_config.env_variables')}</h3>
                  
                  <div class="form-control mb-4">
                    <div class="join">
                      <input
                        type="text"
                        class="input input-bordered join-item"
                        value={envKey()}
                        onInput={(e) => setEnvKey(e.currentTarget.value)}
                        placeholder={t('cli_project_config.env_key_placeholder')}
                      />
                      <input
                        type="text"
                        class="input input-bordered join-item flex-1"
                        value={envValue()}
                        onInput={(e) => setEnvValue(e.currentTarget.value)}
                        placeholder={t('cli_project_config.env_value_placeholder')}
                      />
                      <button
                        type="button"
                        class="btn btn-primary join-item"
                        onClick={addEnvVar}
                        disabled={!envKey().trim() || !envValue().trim()}
                      >
                        {t('cli_project_config.add_button')}
                      </button>
                    </div>
                  </div>

                  <div class="space-y-2">
                    <For each={Object.entries(config().env)}>{([key, value]) => (
                      <div class="flex items-center gap-2">
                        <span class="font-mono text-sm bg-base-200 px-2 py-1 rounded">{key}={value}</span>
                        <button
                          type="button"
                          class="btn btn-error btn-sm"
                          onClick={() => removeEnvVar(key)}
                        >
                          {t('cli_project_config.remove_button')}
                        </button>
                      </div>
                    )}</For>
                  </div>
                </div>

                {/* Volume Mounts */}
                <div class="mt-6">
                  <div class="flex justify-between items-center mb-4">
                    <h3 class="font-semibold text-lg">{t('cli_project_config.persistent_storage')}</h3>
                    <button
                      type="button"
                      class="btn btn-outline btn-sm"
                      onClick={addVolumeMount}
                    >
                      {t('cli_project_config.add_mount')}
                    </button>
                  </div>

                  <div class="space-y-3">
                    <For each={config().volumes}>{(volume, index) => (
                      <div class="grid grid-cols-12 gap-2 items-end">
                        <div class="col-span-5">
                          <label class="label label-text-sm">{t('cli_project_config.host_path')}</label>
                          <input
                            type="text"
                            class="input input-bordered input-sm w-full"
                            value={volume.hostPath}
                            onInput={(e) => updateVolumeMount(index(), 'hostPath', e.currentTarget.value)}
                            placeholder={t('cli_project_config.host_path_placeholder')}
                          />
                        </div>
                        <div class="col-span-5">
                          <label class="label label-text-sm">{t('cli_project_config.container_path')}</label>
                          <input
                            type="text"
                            class="input input-bordered input-sm w-full"
                            value={volume.containerPath}
                            onInput={(e) => updateVolumeMount(index(), 'containerPath', e.currentTarget.value)}
                            placeholder={t('cli_project_config.container_path_placeholder')}
                          />
                        </div>
                        <div class="col-span-2">
                          <button
                            type="button"
                            class="btn btn-error btn-sm w-full"
                            onClick={() => removeVolumeMount(index())}
                          >
                            {t('cli_project_config.remove_button')}
                          </button>
                        </div>
                      </div>
                    )}</For>
                  </div>
                </div>

                {/* Submit Button */}
                <div class="card-actions justify-end mt-8">
                  <button
                    type="button"
                    class="btn btn-ghost"
                    onClick={() => window.close()}
                  >
                    {t('cli_project_config.cancel')}
                  </button>
                  <button
                    type="button"
                    class="btn btn-primary"
                    onClick={handleSubmit}
                    disabled={submitMutation.isPending || expired() || !config().name.trim()}
                  >
                    {submitMutation.isPending && <span class="loading loading-spinner loading-sm"></span>}
                    {submitMutation.isPending ? t('cli_project_config.submitting') : t('cli_project_config.submit_config')}
                  </button>
                </div>
              </div>
            </div>
          </>
        ) : (
          <div class="card bg-base-100 shadow-xl">
            <div class="card-body text-center">
              {submitted() ? (
                <>
                  <div class="text-success text-6xl mb-4">✓</div>
                  <h3 class="text-lg font-semibold text-success">{t('cli_project_config.config_submitted')}</h3>
                  <p class="text-base-content/70">
                    {t('cli_project_config.config_submitted_message')}
                  </p>
                  <p class="text-sm text-base-content/50 mt-2">
                    {t('cli_project_config.auto_close_notice')}
                  </p>
                </>
              ) : (
                <>
                  <div class="text-error text-6xl mb-4">⏰</div>
                  <h3 class="text-lg font-semibold text-error">{t('cli_project_config.session_expired')}</h3>
                  <p class="text-base-content/70">
                    {t('cli_project_config.session_expired_message')}
                  </p>
                  <button
                    class="btn btn-ghost mt-4"
                    onClick={() => window.close()}
                  >
                    {t('cli_project_config.close_page')}
                  </button>
                </>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default CLIProjectConfigPage