import { createSignal } from 'solid-js'
import type { Component } from 'solid-js'
import { useApiMutation } from '../api/apiHooksW.ts'
import { createGithubTokenEndpoint } from '../api/endpoints/githubTokens'

interface CreateTokenModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
}

const CreateTokenModal: Component<CreateTokenModalProps> = (props) => {

  const [formData, setFormData] = createSignal({
    name: '',
    token: '',
    permissions: 'repo',
    expires_at: ''
  })

  const [isValidating, setIsValidating] = createSignal(false)
  const [validationResult, setValidationResult] = createSignal<{
    valid: boolean
    message: string
  } | null>(null)

  // 创建令牌
  const createMutation = useApiMutation<unknown, any>(
    createGithubTokenEndpoint(),
    {
      onSuccess: () => {
        props.onClose()
        resetForm()
        props.onSuccess('GitHub令牌创建成功')
      },
      onError: (err: Error) => {
        props.onError(err.message)
      }
    }
  )

  const resetForm = () => {
    setFormData({
      name: '',
      token: '',
      permissions: 'repo',
      expires_at: ''
    })
    setValidationResult(null)
  }

  const validateToken = async () => {
    const token = formData().token.trim()
    if (!token) {
      setValidationResult({
        valid: false,
        message: '请输入GitHub令牌'
      })
      return
    }

    if (!token.startsWith('ghp_') && !token.startsWith('github_pat_')) {
      setValidationResult({
        valid: false,
        message: '令牌格式无效，应以 ghp_ 或 github_pat_ 开头'
      })
      return
    }

    setIsValidating(true)
    setValidationResult(null)

    try {
      // 这里可以添加实际的GitHub API验证逻辑
      // 暂时进行基本格式验证
      await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用

      setValidationResult({
        valid: true,
        message: '令牌格式验证通过'
      })
    } catch (error) {
      setValidationResult({
        valid: false,
        message: '令牌验证失败，请检查令牌是否有效'
      })
    } finally {
      setIsValidating(false)
    }
  }

  const handleSubmit = () => {
    const form = formData()
    
    if (!form.name.trim()) {
      props.onError('请输入令牌名称')
      return
    }

    if (!form.token.trim()) {
      props.onError('请输入GitHub令牌')
      return
    }

    const payload: any = {
      name: form.name.trim(),
      token: form.token.trim(),
      permissions: form.permissions
    }

    if (form.expires_at) {
      // Convert local datetime (YYYY-MM-DDTHH:mm) to RFC3339 (UTC) for backend parsing
      const d = new Date(form.expires_at)
      if (!isNaN(d.getTime())) {
        payload.expires_at = d.toISOString()
      }
    }

    createMutation.mutate(payload)
  }

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`}>
      <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg mb-4">添加GitHub访问令牌</h3>
        
        <div class="grid grid-cols-1 gap-4">
          <div>
            <label class="label">
              <span class="label-text">令牌名称 *</span>
            </label>
            <input
              class="input input-bordered w-full"
              value={formData().name}
              onInput={(e) => setFormData(prev => ({ ...prev, name: e.currentTarget.value }))}
              placeholder="例如：我的私有仓库令牌"
            />
          </div>

          <div>
            <label class="label">
              <span class="label-text">GitHub Personal Access Token *</span>
            </label>
            <div class="form-control">
              <textarea
                class="textarea textarea-bordered h-20"
                value={formData().token}
                onInput={(e) => setFormData(prev => ({ ...prev, token: e.currentTarget.value }))}
                placeholder="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
              />
              <div class="label">
                <span class="label-text-alt text-gray-500">
                  请输入您的GitHub Personal Access Token
                </span>
              </div>
            </div>
            
            {/* 令牌验证 */}
            <div class="mt-2">
              <button
                class="btn btn-sm btn-outline"
                disabled={isValidating() || !formData().token.trim()}
                onClick={validateToken}
              >
                {isValidating() ? '验证中...' : '验证令牌格式'}
              </button>
              
              {validationResult() && (
                <div class={`mt-2 text-sm ${validationResult()!.valid ? 'text-success' : 'text-error'}`}>
                  {validationResult()!.message}
                </div>
              )}
            </div>
          </div>

          <div>
            <label class="label">
              <span class="label-text">权限范围</span>
            </label>
            <select
              class="select select-bordered w-full"
              value={formData().permissions}
              onInput={(e) => setFormData(prev => ({ ...prev, permissions: e.currentTarget.value }))}
            >
              <option value="repo">repo - 完整仓库访问权限</option>
              <option value="repo:status">repo:status - 仓库状态访问</option>
              <option value="repo_deployment">repo_deployment - 部署访问</option>
              <option value="public_repo">public_repo - 公开仓库访问</option>
            </select>
            <div class="label">
              <span class="label-text-alt text-gray-500">
                推荐选择 "repo" 以获得私有仓库完整访问权限
              </span>
            </div>
          </div>

          <div>
            <label class="label">
              <span class="label-text">过期时间 (可选)</span>
            </label>
            <input
              class="input input-bordered w-full"
              type="datetime-local"
              value={formData().expires_at}
              onInput={(e) => setFormData(prev => ({ ...prev, expires_at: e.currentTarget.value }))}
            />
            <div class="label">
              <span class="label-text-alt text-gray-500">
                留空表示不设置过期时间
              </span>
            </div>
          </div>
        </div>

        {/* 创建指南 */}
        <div class="bg-base-200 rounded-lg p-4 mt-4">
          <h4 class="font-semibold mb-2">如何创建GitHub Personal Access Token：</h4>
          <ol class="list-decimal list-inside text-sm space-y-1">
            <li>访问GitHub → Settings → Developer settings</li>
            <li>选择 Personal access tokens → Tokens (classic)</li>
            <li>点击 "Generate new token" → "Generate new token (classic)"</li>
            <li>设置Token描述和过期时间</li>
            <li>勾选 "repo" 权限范围（用于私有仓库访问）</li>
            <li>点击 "Generate token" 并复制生成的令牌</li>
          </ol>
          <div class="mt-2">
            <a
              href="https://github.com/settings/tokens"
              target="_blank"
              class="link link-primary text-sm"
            >
              直接前往GitHub创建令牌 ↗
            </a>
          </div>
        </div>

        <div class="modal-action">
          <button
            class="btn btn-primary"
            disabled={createMutation.isPending || !formData().name.trim() || !formData().token.trim()}
            onClick={handleSubmit}
          >
            {createMutation.isPending ? '创建中...' : '创建令牌'}
          </button>
          <button
            class="btn"
            onClick={() => {
              props.onClose()
              resetForm()
            }}
          >
            取消
          </button>
        </div>
      </div>
    </div>
  )
}

export default CreateTokenModal