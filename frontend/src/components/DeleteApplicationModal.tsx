import { createSignal, Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import type { Application } from '../types/project'

interface DeleteApplicationModalProps {
  isOpen: boolean
  application: Application | null
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
  onRefresh: () => void
}

const DeleteApplicationModal: Component<DeleteApplicationModalProps> = (props) => {
  const { t } = useI18n()
  const [deleteConfirmName, setDeleteConfirmName] = createSignal('')
  const [error, setError] = createSignal('')
  const [isDeleting, setIsDeleting] = createSignal(false)

  async function confirmDelete() {
    const application = props.application
    if (!application) return
    
    setError('')
    
    // Verify application name matches
    if (deleteConfirmName().trim() !== application.name) {
      setError('应用名称确认不匹配')
      return
    }
    
    setIsDeleting(true)
    
    try {
      const res = await fetch(`/api/apps/${application.uid}`, {
        method: 'DELETE',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`
        },
        body: JSON.stringify({
          applicationName: application.name
        })
      })
      
      const json = await res.json()
      
      if (!res.ok || !json?.success) {
        const msg = (json && json.message) ? json.message : '删除应用失败'
        throw new Error(msg)
      }
      
      props.onClose()
      setDeleteConfirmName('')
      setError('')
      props.onSuccess('应用删除成功')
      props.onRefresh()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setIsDeleting(false)
    }
  }

  function handleClose() {
    if (isDeleting()) return // Prevent closing while deleting
    setDeleteConfirmName('')
    setError('')
    props.onClose()
  }

  return (
    <div class={`modal ${props.isOpen && props.application ? 'modal-open' : ''}`}>
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4 text-error">删除应用</h3>
        <div class="space-y-4">
          <div class="alert alert-warning">
            <div>
              <strong>警告：</strong> 此操作无法撤销。这将会：
              <ul class="list-disc list-inside mt-2 space-y-1">
                <li>停止并删除所有相关的容器</li>
                <li>删除所有域名关联</li>
                <li>删除通过此应用构建的所有镜像</li>
                <li>删除所有构建和部署历史记录</li>
                <li>删除所有配置文件和本地文件</li>
              </ul>
            </div>
          </div>
          <div>
            <p class="mb-2">
              请输入 <strong>{props.application?.name || ''}</strong> 来确认删除：
            </p>
            <input 
              class="input input-bordered w-full" 
              value={deleteConfirmName()} 
              onInput={(e) => setDeleteConfirmName(e.currentTarget.value)}
              placeholder="输入应用名称"
              disabled={isDeleting()}
            />
          </div>
          <Show when={error()}>
            <div class="alert alert-error">
              <div>{error()}</div>
            </div>
          </Show>
        </div>
        <div class="modal-action">
          <button 
            class="btn btn-error" 
            disabled={!props.application || deleteConfirmName().trim() !== props.application.name || isDeleting()}
            onClick={() => void confirmDelete()}
          >
            <Show when={isDeleting()} fallback="删除应用">
              <span class="loading loading-spinner loading-sm"></span>
              删除中...
            </Show>
          </button>
          <button 
            class="btn" 
            disabled={isDeleting()}
            onClick={handleClose}
          >
            {t('common.cancel') || '取消'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default DeleteApplicationModal