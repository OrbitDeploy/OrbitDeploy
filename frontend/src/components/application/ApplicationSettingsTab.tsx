import { Component, Show, For, createSignal, createEffect } from 'solid-js'
import type { Application, VolumeMount } from '../../types/project'
import { useI18n } from '../../i18n'
import { useApiMutation } from '../../api/apiHooksW.ts'
import { updateApplicationEndpoint } from '../../api/endpoints'
import { useNavigate } from '@solidjs/router'
import DeleteApplicationModal from '../DeleteApplicationModal'

interface ApplicationSettingsTabProps {
  currentApp: Application | undefined
}

const ApplicationSettingsTab: Component<ApplicationSettingsTabProps> = (props) => {
  const { t } = useI18n()
  const navigate = useNavigate()
  
  // 本地状态管理
  const [volumeMounts, setVolumeMounts] = createSignal<VolumeMount[]>([])
  const [execCommand, setExecCommand] = createSignal<string>('')
  const [autoUpdatePolicy, setAutoUpdatePolicy] = createSignal<string>('')
  const [description, setDescription] = createSignal<string>('')
  const [targetPort, setTargetPort] = createSignal<number | undefined>()
  const [branch, setBranch] = createSignal<string>('')
  const [repoUrl, setRepoUrl] = createSignal<string>('')
  const [buildDir, setBuildDir] = createSignal<string>('')
  const [buildType, setBuildType] = createSignal<string>('')
  const [providerAuthId, setProviderAuthId] = createSignal<number | undefined>()
  const [isSaving, setIsSaving] = createSignal(false)
  const [error, setError] = createSignal('')
  const [successMessage, setSuccessMessage] = createSignal('')
  const [showDeleteModal, setShowDeleteModal] = createSignal(false)

  // 从当前应用初始化状态
  createEffect(() => {
    if (props.currentApp) {
      // 正确处理 volumes 数据，确保它是 VolumeMount 数组
      let volumes: VolumeMount[] = []
      if (Array.isArray(props.currentApp.volumes)) {
        volumes = props.currentApp.volumes.map((vol: any) => ({
          hostPath: vol.hostPath || vol.host_path || '',
          containerPath: vol.containerPath || vol.container_path || '',
          readOnly: vol.readOnly || vol.read_only || false
        }))
      }
      setVolumeMounts(volumes)
      
      setExecCommand(props.currentApp.execCommand || '')
      setAutoUpdatePolicy(props.currentApp.autoUpdatePolicy || '')
      setDescription(props.currentApp.description || '')
      setTargetPort(props.currentApp.targetPort)
      setBranch(props.currentApp.branch || 'main')
      setRepoUrl(props.currentApp.repoUrl || '')
      setBuildDir(props.currentApp.buildDir || '/')
      setBuildType(props.currentApp.buildType || 'dockerfile')
    }
  })

  // 更新应用配置的 API 调用
  const updateAppMutation = useApiMutation<any, any>(
    () => props.currentApp ? updateApplicationEndpoint(props.currentApp.uid) : null,
    {
      onSuccess: () => {
        setSuccessMessage('应用配置已成功更新')
        setError('')
        // 3秒后清除成功消息
        setTimeout(() => setSuccessMessage(''), 3000)
      },
      onError: (error: any) => {
        setError(error?.message || '保存失败，请重试')
        setSuccessMessage('')
      }
    }
  )

  // 卷挂载管理函数
  const addVolumeMount = () => {
    setVolumeMounts(prev => [...prev, { hostPath: '', containerPath: '', readOnly: false }])
  }

  const removeVolumeMount = (index: number) => {
    setVolumeMounts(prev => prev.filter((_, i) => i !== index))
  }

  const updateVolumeMount = (index: number, field: keyof VolumeMount, value: string | boolean) => {
    setVolumeMounts(prev => prev.map((mount, i) => 
      i === index ? { ...mount, [field]: value } : mount
    ))
  }

  // Debounce utility function
  const debounce = (func: Function, delay: number) => {
    let timeoutId: ReturnType<typeof setTimeout>;
    return (...args: any[]) => {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => func.apply(null, args), delay);
    };
  };

  // Debounced update functions
  const debouncedUpdateVolumeMount = debounce(updateVolumeMount, 300);

  // 通用保存函数，支持部分更新
  const handlePartialSave = async (updateData: Partial<any>) => {
    if (!props.currentApp) return

    setIsSaving(true)
    setError('')

    try {
      await updateAppMutation.mutateAsync(updateData)
    } catch (err: any) {
      setError(err.message || '保存失败，请重试')
    } finally {
      setIsSaving(false)
    }
  }

  // 单独保存函数
  const handleSaveBasicInfo = () => handlePartialSave({
    description: description().trim(),
    targetPort: targetPort()
  })

  const handleSaveRepoConfig = () => handlePartialSave({
    repoUrl: repoUrl().trim(),
    branch: branch().trim(),
    buildDir: buildDir().trim(),
    buildType: buildType().trim(),
    providerAuthId: providerAuthId()
  })

  const handleSaveRuntime = () => handlePartialSave({
    execCommand: execCommand().trim() || null,
    autoUpdatePolicy: autoUpdatePolicy() || null
  })

  const handleSaveStorage = () => handlePartialSave({
    volumes: volumeMounts()
  })

  // 处理删除成功后的导航
  const handleDeleteSuccess = (message: string) => {
    // 导航回到项目详情页
    if (props.currentApp?.projectUid) {
      navigate(`/projects/${props.currentApp.projectUid}`)
    } else {
      navigate('/projects')
    }
  }

  // 处理删除错误
  const handleDeleteError = (message: string) => {
    setError(message)
  }

  return (
    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold">应用设置</h3>
      </div>

      {/* 消息提示 */}
      <Show when={error()}>
        <div class="alert alert-error">
          <span>{error()}</span>
        </div>
      </Show>
      
      <Show when={successMessage()}>
        <div class="alert alert-success">
          <span>{successMessage()}</span>
        </div>
      </Show>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Basic Information */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <div class="flex items-center justify-between">
              <h4 class="card-title">基本信息</h4>
              <button 
                class="btn btn-outline btn-sm" 
                onClick={handleSaveBasicInfo}
                disabled={isSaving() || !props.currentApp}
              >
                保存
              </button>
            </div>
            <div class="space-y-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">应用描述</span>
                </label>
                <textarea
                  class="textarea textarea-bordered"
                  value={description()}
                  onInput={(e) => setDescription(e.currentTarget.value)}
                  placeholder="输入应用描述"
                  rows="3"
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">目标端口 *</span>
                </label>
                <input
                  type="number"
                  class="input input-bordered"
                  min="1"
                  max="65535"
                  value={targetPort() ?? ''}
                  onInput={(e) => {
                    const val = e.currentTarget.value
                    setTargetPort(val === '' ? undefined : parseInt(val))
                  }}
                  placeholder="容器内部监听的端口号"
                />
              </div>
            </div>
          </div>
        </div>

        {/* Repository Configuration */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <div class="flex items-center justify-between">
              <h4 class="card-title">仓库配置</h4>
              <button 
                class="btn btn-outline btn-sm" 
                onClick={handleSaveRepoConfig}
                disabled={isSaving() || !props.currentApp}
              >
                保存
              </button>
            </div>
            <div class="space-y-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">仓库URL</span>
                </label>
                <input
                  type="text"
                  class="input input-bordered"
                  value={repoUrl()}
                  onInput={(e) => setRepoUrl(e.currentTarget.value)}
                  placeholder="https://github.com/user/repo"
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">分支</span>
                </label>
                <input
                  type="text"
                  class="input input-bordered"
                  value={branch()}
                  onInput={(e) => setBranch(e.currentTarget.value)}
                  placeholder="输入分支名称"
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">构建目录</span>
                </label>
                <input
                  type="text"
                  class="input input-bordered"
                  value={buildDir()}
                  onInput={(e) => setBuildDir(e.currentTarget.value)}
                  placeholder="构建目录，默认根目录"
                />
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">构建类型</span>
                </label>
                <select 
                  class="select select-bordered"
                  value={buildType()}
                  onChange={(e) => setBuildType(e.currentTarget.value)}
                >
                  <option value="dockerfile">Dockerfile</option>
                  {/* <option value="railpack">Railpack</option>
                  <option value="nixpacks">Nixpacks</option> */}
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Runtime Configuration */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <div class="flex items-center justify-between">
              <h4 class="card-title">运行时配置</h4>
              <button 
                class="btn btn-outline btn-sm" 
                onClick={handleSaveRuntime}
                disabled={isSaving() || !props.currentApp}
              >
                保存
              </button>
            </div>
            <div class="space-y-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">执行命令</span>
                </label>
                <input
                  type="text"
                  class="input input-bordered"
                  value={execCommand()}
                  onInput={(e) => setExecCommand(e.currentTarget.value)}
                  placeholder="可选的容器启动命令，例如：/start.sh"
                />
                <div class="label">
                  <span class="label-text-alt">覆盖镜像的默认启动命令</span>
                </div>
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">自动更新策略</span>
                </label>
                <select 
                  class="select select-bordered"
                  value={autoUpdatePolicy()}
                  onChange={(e) => setAutoUpdatePolicy(e.currentTarget.value)}
                >
                  <option value="">手动更新</option>
                  <option value="registry">镜像仓库更新</option>
                </select>
                <div class="label">
                  <span class="label-text-alt">选择自动更新模式</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Volume Configuration */}
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <div class="flex items-center justify-between">
              <h4 class="card-title">存储配置</h4>
              <button 
                class="btn btn-outline btn-sm" 
                onClick={handleSaveStorage}
                disabled={isSaving() || !props.currentApp}
              >
                保存
              </button>
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('projects.form.persistent_storage')}</span>
              </label>
              <button 
                class="btn btn-outline btn-sm"
                onClick={addVolumeMount}
              >
                {t('projects.form.add_mount')}
              </button>
            </div>
            
            <Show when={volumeMounts().length > 0} fallback={
              <div class="text-center text-base-content/70 py-8">
                <p>暂无卷挂载配置</p>
                <p class="text-sm mt-2">点击"添加卷挂载"按钮开始配置</p>
              </div>
            }>
              <div class="space-y-3">
                <For each={volumeMounts()}>{(mount, index) => (
                  <div class="border border-base-300 rounded-lg p-4 space-y-3">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-medium">卷挂载 #{index() + 1}</span>
                      <button 
                        type="button"
                        class="btn btn-sm btn-error btn-outline"
                        onClick={() => removeVolumeMount(index())}
                        title={t('projects.form.remove_mount')}
                      >
                        🗑️
                      </button>
                    </div>
                    
                    <div class="grid grid-cols-1 gap-3">
                      <div class="form-control">
                        <label class="label label-text-sm">{t('projects.form.host_path')}</label>
                        <input 
                          type="text"
                          class="input input-bordered input-sm"
                          value={mount.hostPath}
                          onInput={(e) => debouncedUpdateVolumeMount(index(), 'hostPath', e.currentTarget.value)}
                          placeholder={t('projects.form.host_path_placeholder')}
                        />
                        <div class="label">
                          <span class="label-text-alt">相对于服务器数据目录的路径</span>
                        </div>
                      </div>
                      
                      <div class="form-control">
                        <label class="label label-text-sm">{t('projects.form.container_path')}</label>
                        <input 
                          type="text"
                          class="input input-bordered input-sm"
                          value={mount.containerPath}
                          onInput={(e) => debouncedUpdateVolumeMount(index(), 'containerPath', e.currentTarget.value)}
                          placeholder={t('projects.form.container_path_placeholder')}
                        />
                        <div class="label">
                          <span class="label-text-alt">容器内的挂载点，必须以 / 开头</span>
                        </div>
                      </div>

                      <div class="form-control">
                        <label class="label cursor-pointer">
                          <span class="label-text">只读模式</span>
                          <input 
                            type="checkbox" 
                            class="checkbox checkbox-primary"
                            checked={mount.readOnly || false}
                            onChange={(e) => updateVolumeMount(index(), 'readOnly', e.currentTarget.checked)}
                          />
                        </label>
                        <div class="label">
                          <span class="label-text-alt">启用后容器无法修改主机文件</span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}</For>
              </div>
            </Show>
            
            <div class="alert alert-info mt-4 break-words">
              <div class="text-sm">
                <p><strong>💡 卷挂载说明：</strong></p>
                <ul class="list-disc list-inside mt-2 space-y-1">
                  <li><strong>基本目录挂载：</strong> Volume=/var/app/data:/app/data</li>
                  <li><strong>只读挂载：</strong> Volume=/var/app/config:/app/config:ro</li>
                  <li><strong>命名卷：</strong> Volume=postgres_data:/var/lib/postgresql/data</li>
                </ul>
                <p class="mt-2">数据将持久保存在服务器上，确保容器重启后数据不丢失。</p>
              </div>
            </div>
          </div>
        </div>

        {/* Danger Zone */}
        <div class="card bg-base-100 shadow-xl border-error">
          <div class="card-body">
            <h4 class="card-title text-error">危险操作</h4>
            <div class="flex gap-2">
              <button class="btn btn-outline btn-error btn-sm">重置应用</button>
              <button 
                class="btn btn-error btn-sm"
                onClick={() => setShowDeleteModal(true)}
                disabled={!props.currentApp}
              >
                删除应用
              </button>
            </div>
            <p class="text-sm text-base-content/70">这些操作不可撤销，请谨慎使用。</p>
          </div>
        </div>
      </div>

      {/* Delete Application Modal */}
      <DeleteApplicationModal
        isOpen={showDeleteModal()}
        application={props.currentApp || null}
        onClose={() => setShowDeleteModal(false)}
        onSuccess={handleDeleteSuccess}
        onError={handleDeleteError}
        onRefresh={() => {}} // Not needed for settings tab since we navigate away
      />
    </div>
  )
}

export default ApplicationSettingsTab