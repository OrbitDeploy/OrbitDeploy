import { createSignal, onMount, createMemo } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import type { GitHubToken } from '../types/githubToken'
import { useApiQuery, useApiMutation } from '../lib/apiHooks'
import { getGitHubTokensApiUrl } from '../api/config'
import CreateTokenModal from '../components/CreateTokenModal'

const GitHubTokenManagementPage: Component = () => {
  const { t } = useI18n()
  const [isCreateModalOpen, setIsCreateModalOpen] = createSignal(false)
  const [message, setMessage] = createSignal('')
  const [messageType, setMessageType] = createSignal<'success' | 'error'>('success')

  // 获取GitHub令牌列表
  const tokensQuery = useApiQuery<GitHubToken[]>(
    ['github-tokens'],
    () => getGitHubTokensApiUrl('list')
  )

  // 删除令牌
  const deleteMutation = useApiMutation<unknown, { id: number }>(
    ({ id }) => getGitHubTokensApiUrl({ type: 'delete', id }),
    {
      method: 'DELETE',
      onSuccess: () => {
        setMessage('GitHub令牌删除成功')
        setMessageType('success')
        tokensQuery.refetch()
      },
      onError: (err) => {
        setMessage(`删除失败: ${err.message}`)
        setMessageType('error')
      }
    }
  )

  // Define the type for the test token response
  type TestTokenResponse = {
    success: boolean
    data: {
      valid: boolean
      permissions: string[]
      username: string
      rate_limit: {
        remaining: number
        total: number
        reset_at: number
      }
    }
  }

  // 测试令牌
  const testMutation = useApiMutation<TestTokenResponse, { id: number }>(
    ({ id }) => getGitHubTokensApiUrl({ type: 'test', id }),
    {
      method: 'POST',
      onSuccess: (response) => {
        if (response.data.valid) {
          setMessage(`令牌验证成功，用户: ${response.data.username}`)
          setMessageType('success')
        } else {
          setMessage('令牌无效或已过期')
          setMessageType('error')
        }
      },
      onError: (err) => {
        setMessage(`令牌测试失败: ${err.message}`)
        setMessageType('error')
      }
    }
  )

  const showMessage = (msg: string, type: 'success' | 'error') => {
    setMessage(msg)
    setMessageType(type)
    setTimeout(() => setMessage(''), 5000)
  }

  const handleDelete = (token: GitHubToken) => {
    if (confirm(`确定要删除令牌 "${token.name}" 吗？此操作不可撤销。`)) {
      deleteMutation.mutate({ id: token.uid })
    }
  }

  const handleTest = (token: GitHubToken) => {
    testMutation.mutate({ id: token.uid })
  }

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '从未使用'
    return new Date(dateStr).toLocaleString('zh-CN')
  }

  const tokens = createMemo(() => tokensQuery.data || [])
  const isLoading = createMemo(() => tokensQuery.isPending)
  const error = createMemo(() => tokensQuery.error)

  return (
    <div class="container mx-auto p-6">
      <div class="flex justify-between items-center mb-6">
        <div>
          <h1 class="text-2xl font-bold">GitHub 访问令牌管理</h1>
          <p class="text-gray-600 mt-2">
            管理用于访问私有GitHub仓库的Personal Access Token
          </p>
        </div>
        <button
          class="btn btn-primary"
          onClick={() => setIsCreateModalOpen(true)}
        >
          添加令牌
        </button>
      </div>

      {/* 消息提示 */}
      {message() && (
        <div class={`alert mb-4 ${messageType() === 'success' ? 'alert-success' : 'alert-error'}`}>
          <span>{message()}</span>
        </div>
      )}

      {/* 使用指南 */}
      <div class="card bg-base-200 mb-6">
        <div class="card-body">
          <h2 class="card-title text-lg">使用指南</h2>
          <div class="text-sm space-y-2">
            <p>
              <strong>1. 创建GitHub Personal Access Token:</strong>
              访问 GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
            </p>
            <p>
              <strong>2. 推荐权限范围:</strong>
              repo (完整仓库访问权限) 用于私有仓库拉取
            </p>
            <p>
              <strong>3. 安全提示:</strong>
              令牌将被加密存储，仅在构建时使用。请定期轮换令牌以保证安全。
            </p>
          </div>
        </div>
      </div>

      {/* 令牌列表 */}
      {isLoading() && (
        <div class="flex justify-center py-8">
          <span class="loading loading-spinner loading-md"></span>
        </div>
      )}

      {error() && (
        <div class="alert alert-error">
          <span>加载失败: {error() instanceof Error ? error().message : '未知错误'}</span>
        </div>
      )}

      {tokens().length === 0 && !isLoading() && (
        <div class="card bg-base-100">
          <div class="card-body text-center">
            <h3 class="text-lg font-semibold mb-2">暂无GitHub令牌</h3>
            <p class="text-gray-600 mb-4">
              添加您的第一个GitHub Personal Access Token以访问私有仓库
            </p>
            <button
              class="btn btn-primary"
              onClick={() => setIsCreateModalOpen(true)}
            >
              添加令牌
            </button>
          </div>
        </div>
      )}

      {tokens().length > 0 && (
        <div class="grid gap-4">
          {tokens().map((token) => (
            <div class="card bg-base-100 shadow">
              <div class="card-body">
                <div class="flex justify-between items-start">
                  <div class="flex-1">
                    <h3 class="card-title text-lg">{token.name}</h3>
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4 text-sm">
                      <div>
                        <span class="text-gray-500">权限范围:</span>
                        <div class="mt-1">
                          {token.permissions ? (
                            <span class="badge badge-outline">{token.permissions}</span>
                          ) : (
                            <span class="text-gray-400">未设置</span>
                          )}
                        </div>
                      </div>
                      <div>
                        <span class="text-gray-500">最后使用:</span>
                        <div class="mt-1">{formatDate(token.last_used_at)}</div>
                      </div>
                      <div>
                        <span class="text-gray-500">创建时间:</span>
                        <div class="mt-1">{formatDate(token.created_at)}</div>
                      </div>
                    </div>
                    {token.expires_at && (
                      <div class="mt-2 text-sm">
                        <span class="text-gray-500">过期时间:</span>
                        <span class="ml-2">{formatDate(token.expires_at)}</span>
                      </div>
                    )}
                  </div>
                  <div class={`badge ${token.is_active ? 'badge-success' : 'badge-error'}`}>
                    {token.is_active ? '激活' : '禁用'}
                  </div>
                </div>
                <div class="card-actions justify-end mt-4">
                  <button
                    class="btn btn-sm btn-outline"
                    disabled={testMutation.isPending}
                    onClick={() => handleTest(token)}
                  >
                    {testMutation.isPending ? '测试中...' : '测试令牌'}
                  </button>
                  <button
                    class="btn btn-sm btn-error"
                    disabled={deleteMutation.isPending}
                    onClick={() => handleDelete(token)}
                  >
                    删除
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 创建令牌模态框 */}
      <CreateTokenModal
        isOpen={isCreateModalOpen()}
        onClose={() => setIsCreateModalOpen(false)}
        onSuccess={(message) => {
          showMessage(message, 'success')
          tokensQuery.refetch()
        }}
        onError={(message) => showMessage(message, 'error')}
      />
    </div>
  )
}

export default GitHubTokenManagementPage