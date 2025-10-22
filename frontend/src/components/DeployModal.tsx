import { createSignal, Show, For, createEffect } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import type { 
  Project, 
  ImageInfo, 
  VolumeMount, 
  Environment 
} from '../types/project'
import { 
  parseDomainsString, 
  fetchEnvironments 
} from '../services/projectService'
import { useApiQuery, useApiMutation } from '../lib/apiHooks'
import { getProjectsApiUrl } from '../api/config'

interface DeployModalProps {
  isOpen: boolean
  project: Project | null
  selectedEnvironment: string  // 新增：从父组件传入的选中环境
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
  onRefreshDeployments: () => void
  onShowLogs: (deploymentId: number | null, containerId?: number) => void
}

const DeployModal: Component<DeployModalProps> = (props) => {
  const { t } = useI18n()

  // 表单状态
  const [imageRef, setImageRef] = createSignal('')
  const [publishPort, setPublishPort] = createSignal<number>(8080)
  const [replicas, setReplicas] = createSignal<number>(1)
  const [strategy, setStrategy] = createSignal('direct')
  const [domains, setDomains] = createSignal('')
  const [envFile, setEnvFile] = createSignal('')
  const [volumeMounts, setVolumeMounts] = createSignal<VolumeMount[]>([])
  const [enableRebuild, setEnableRebuild] = createSignal(false) // 新增：是否启用重新构建
  
  // UI 状态
  const [submitError, setSubmitError] = createSignal('')

  // 使用 useApiQuery 获取可用镜像
  const availableImagesQuery = useApiQuery<ImageInfo[]>(
    () => ['project-images', props.project?.uid],
    () => props.project ? getProjectsApiUrl({ type: 'images', uid: props.project.uid }) : null,
    {
      enabled: () => props.isOpen && !!props.project
    }
  )

  // 使用 useApiMutation 处理部署请求
  const deployMutation = useApiMutation<any, any>(
    () => props.project ? getProjectsApiUrl({ type: 'deploy', uid: props.project.uid }) : '',
    {
      method: 'POST',
      onSuccess: (data: any) => {
        props.onClose()
        props.onSuccess('部署已启动，请查看日志了解详细进度')
        props.onRefreshDeployments()

        // 如果返回了 deployment_id，打开日志
        const deploymentId = data?.deployment_id
        if (deploymentId) {
          props.onShowLogs(deploymentId)
        }
      },
      onError: (err: Error) => {
        const errorMessage = err.message || '部署失败'
        setSubmitError(errorMessage)
        props.onError(errorMessage)
      }
    }
  )

  // 加载环境列表 (keeping for potential future use)
  const loadEnvironments = async () => {
    try {
      const envs = await fetchEnvironments()
      // Currently not used but kept for potential future functionality
      console.log('Loaded environments:', envs)
    } catch (err) {
      console.warn('加载环境列表失败:', err)
    }
  }

  // 初始化模态框
  const initializeModal = async (project: Project) => {
    // 重置表单状态
    setImageRef('')
    setPublishPort(project.publish_port || 8080)
    setReplicas(1)
    setStrategy('direct')
    setDomains('')
    setEnvFile(project.env_template || '') // 预填项目的环境变量模板
    setVolumeMounts([])
    setSubmitError('')
    setEnableRebuild(false) // 重置重新构建选项
    
    // 加载环境数据
    await loadEnvironments()
  }

  // 当模态框打开时初始化
  createEffect(() => {
    if (props.isOpen && props.project) {
      void initializeModal(props.project)
    }
  })

  // 添加卷挂载
  const addVolumeMount = () => {
    setVolumeMounts(prev => [...prev, { hostPath: '', containerPath: '', readOnly: false }])
  }

  // 删除卷挂载
  const removeVolumeMount = (index: number) => {
    setVolumeMounts(prev => prev.filter((_, i) => i !== index))
  }

  // 更新卷挂载
  const updateVolumeMount = (index: number, field: keyof VolumeMount, value: string | boolean) => {
    setVolumeMounts(prev => prev.map((mount, i) => 
      i === index ? { ...mount, [field]: value } : mount
    ))
  }

  // 表单验证
  const validateForm = (): string | null => {
    // 如果启用重新构建，则不需要检查镜像引用
    if (!enableRebuild() && !imageRef().trim()) {
      return t('projects.messages.image_ref_required') || '请选择镜像或启用重新构建'
    }
    
    const port = publishPort()
    if (port < 1 || port > 65535) {
      return t('projects.messages.invalid_port') || '端口号必须在 1-65535 范围内'
    }
    
    // 验证域名格式（简单验证）
    const domainsArray = parseDomainsString(domains())
    for (const domain of domainsArray) {
      if (domain && !/^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$/.test(domain)) {
        return t('projects.messages.invalid_domains') || '域名格式无效'
      }
    }
    
    // 验证卷挂载
    for (let i = 0; i < volumeMounts().length; i++) {
      const mount = volumeMounts()[i]
      if (!mount.hostPath.trim()) {
        return `${t('projects.messages.host_path_required') || '主机路径不能为空'} (第${i + 1}项)`
      }
      if (!mount.containerPath.trim()) {
        return `${t('projects.messages.container_path_required') || '容器路径不能为空'} (第${i + 1}项)`
      }
      if (!mount.containerPath.startsWith('/')) {
        return `${t('projects.messages.invalid_container_path') || '容器路径必须是绝对路径'} (第${i + 1}项)`
      }
    }
    
    return null
  }

  // 提交部署
  const handleDeploy = async () => {
    const project = props.project
    if (!project) return

    // 表单验证
    const validationError = validateForm()
    if (validationError) {
      setSubmitError(validationError)
      return
    }

    setSubmitError('')

    // 构建部署请求
    const deployRequest = {
      rebuild: enableRebuild(),
      image_name: enableRebuild() ? '' : imageRef().trim(),
      container_params: {
        domains: domains().trim(),
        env_file: envFile(),
        description: `Deployed from ${project.Name}`,
      }
    }

    console.log('部署请求参数:', deployRequest)

    // 使用 mutation 提交部署请求
    deployMutation.mutate(deployRequest)
  }

  return (
    <div class={`modal ${props.isOpen && props.project ? 'modal-open' : ''}`}>
      <div class="modal-box max-w-4xl max-h-screen overflow-y-auto">
        <h3 class="font-bold text-lg mb-4">
          {t('projects.modals.deploy_title')} - {props.project?.Name || ''}
        </h3>
        
        {submitError() && (
          <div class="alert alert-error mb-4">
            <span>{submitError()}</span>
          </div>
        )}

        <div class="grid grid-cols-1 gap-4">
          {/* 镜像选择或重新构建 */}
          <div class="form-control">
            <label class="label">
              <span class="label-text">镜像配置</span>
            </label>
            
            {/* 重新构建选项 */}
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">重新构建镜像</span>
                <input 
                  type="checkbox" 
                  class="checkbox" 
                  checked={enableRebuild()}
                  onChange={(e) => setEnableRebuild(e.currentTarget.checked)}
                />
              </label>
              <div class="label">
                <span class="label-text-alt">
                  勾选此项将从项目仓库重新构建最新的镜像，而不是使用已有镜像
                </span>
              </div>
            </div>
            
            {/* 镜像选择（当未启用重新构建时显示） */}
            <Show when={!enableRebuild()}>
              <div class="mt-3">
                <label class="label">
                  <span class="label-text">{t('projects.form.image_ref')} *</span>
                </label>
                <Show when={availableImagesQuery.data && availableImagesQuery.data.length > 0} fallback={
                  <input 
                    type="text"
                    class="input input-bordered w-full"
                    value={imageRef()}
                    onInput={(e) => setImageRef(e.currentTarget.value)}
                    placeholder={t('projects.form.image_ref_placeholder')}
                  />
                }>
                  <select 
                    class="select select-bordered w-full"
                    value={imageRef()}
                    onChange={(e) => setImageRef(e.currentTarget.value)}
                  >
                    <option value="">选择镜像...</option>
                    <For each={availableImagesQuery.data || []}>{(image) => (
                      <option value={image.name}>{image.name}</option>
                    )}</For>
                  </select>
                </Show>
              </div>
            </Show>
            
            {/* 重新构建提示 */}
            <Show when={enableRebuild()}>
              <div class="alert alert-info mt-3">
                <span>将从项目仓库 {props.project?.GitRepository} 重新构建镜像</span>
              </div>
            </Show>
          </div>

          {/* 容器端口 */}
          <div class="form-control">
            <label class="label">
              <span class="label-text">容器端口 (Publish Port)</span>
            </label>
            <input 
              type="number"
              class="input input-bordered w-full"
              value={publishPort()}
              onInput={(e) => setPublishPort(parseInt(e.currentTarget.value) || 8080)}
              min="1"
              max="65535"
            />
          </div>

          {/* 副本数量和策略 */}
          <div class="grid grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('projects.form.replicas')}</span>
              </label>
              <input 
                type="number"
                class="input input-bordered w-full"
                value={replicas()}
                onInput={(e) => setReplicas(parseInt(e.currentTarget.value) || 1)}
                min="1"
                max="10"
              />
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('projects.form.strategy')}</span>
              </label>
              <select 
                class="select select-bordered w-full"
                value={strategy()}
                onChange={(e) => setStrategy(e.currentTarget.value)}
              >
                <option value="direct">Direct (直接部署) - 立即替换现有容器，适合开发环境</option>
                <option value="blue-green">Blue-Green (蓝绿部署) - 零停机部署，先启动新版本再切换流量</option>
                <option value="rolling">Rolling (滚动更新) - 逐步替换实例，保持服务可用性</option>
              </select>
              <div class="label">
                <span class="label-text-alt">
                  {strategy() === 'direct' && '直接部署：停止当前容器，立即启动新版本。部署快速但有短暂停机时间。'}
                  {strategy() === 'blue-green' && '蓝绿部署：新版本与旧版本并行运行，通过负载均衡器切换流量，实现零停机部署。'}
                  {strategy() === 'rolling' && '滚动更新：逐个替换容器实例，确保始终有可用的服务实例运行。'}
                </span>
              </div>
            </div>
          </div>

          {/* 域名配置 */}
          <div class="form-control">
            <label class="label">
              <span class="label-text">{t('projects.form.domains')}</span>
            </label>
            <input 
              type="text"
              class="input input-bordered w-full"
              value={domains()}
              onInput={(e) => setDomains(e.currentTarget.value)}
              placeholder="app.example.com, api.example.com"
            />
            <div class="label">
              <span class="label-text-alt">多个域名用逗号分隔</span>
            </div>
          </div>

          {/* 环境变量文件 */}
          <div class="form-control">
            <label class="label">
              <span class="label-text">{t('projects.form.env_file')}</span>
            </label>
            <textarea 
              class="textarea textarea-bordered w-full h-24"
              value={envFile()}
              onInput={(e) => setEnvFile(e.currentTarget.value)}
              placeholder={t('projects.form.env_placeholder')}
            />
          </div>

          {/* 持久化存储配置 */}
          <div class="form-control">
            <label class="label">
              <span class="label-text">{t('projects.form.persistent_storage')}</span>
              <button 
                type="button"
                class="btn btn-sm btn-outline"
                onClick={addVolumeMount}
              >
                {t('projects.form.add_mount')}
              </button>
            </label>
            
            <Show when={volumeMounts().length > 0}>
              <div class="space-y-3">
                <For each={volumeMounts()}>{(mount, index) => (
                  <div class="grid grid-cols-12 gap-2 items-end">
                    <div class="col-span-5">
                      <label class="label label-text-sm">{t('projects.form.host_path')}</label>
                      <input 
                        type="text"
                        class="input input-bordered input-sm w-full"
                        value={mount.hostPath}
                        onInput={(e) => updateVolumeMount(index(), 'hostPath', e.currentTarget.value)}
                        placeholder={t('projects.form.host_path_placeholder')}
                      />
                    </div>
                    <div class="col-span-5">
                      <label class="label label-text-sm">{t('projects.form.container_path')}</label>
                      <input 
                        type="text"
                        class="input input-bordered input-sm w-full"
                        value={mount.containerPath}
                        onInput={(e) => updateVolumeMount(index(), 'containerPath', e.currentTarget.value)}
                        placeholder={t('projects.form.container_path_placeholder')}
                      />
                    </div>
                    <div class="col-span-2">
                      <button 
                        type="button"
                        class="btn btn-sm btn-error w-full"
                        onClick={() => removeVolumeMount(index())}
                        title={t('projects.form.remove_mount')}
                      >
                        🗑️
                      </button>
                    </div>
                  </div>
                )}</For>
              </div>
            </Show>
            
            <div class="label">
              <span class="label-text-alt">
                <span class="tooltip" data-tip={t('projects.form.host_path_help')}>
                  💡 {t('projects.form.host_path_help')}
                </span>
              </span>
            </div>
          </div>
        </div>

        <div class="modal-action">
          <button 
            class="btn btn-primary" 
            disabled={deployMutation.isPending} 
            onClick={handleDeploy}
          >
            {deployMutation.isPending && <span class="loading loading-spinner loading-sm"></span>}
            {deployMutation.isPending
              ? t('common.loading') 
              : enableRebuild() 
                ? '构建并部署' 
                : t('projects.actions.deploy')
            }
          </button>
          <button 
            class="btn" 
            disabled={deployMutation.isPending} 
            onClick={props.onClose}
          >
            {t('common.cancel')}
          </button>
        </div>
      </div>
    </div>
  )
}

export default DeployModal