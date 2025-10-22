import { createSignal } from 'solid-js'
import type { Component } from 'solid-js'
import EnvironmentCheckModal from '../components/EnvironmentCheckModal'


interface SetupPageProps {
  onSetupComplete: () => void
}

// 这是 TypeScript 的接口定义，声明了一个名为 SetupPageProps 的类型，里面有一个属性 onSetupComplete，类型是函数（无参数，返回值为 void）。用于约束组件的 props 类型。

/*
const SetupPage: Component<SetupPageProps> = (props) => {
*/
// 这是 Solid.js 组件的定义方式。Component<SetupPageProps> 表示该组件的 props 类型为 SetupPageProps。props 是组件接收到的参数对象，可以通过 props.onSetupComplete 调用传入的函数。
const SetupPage: Component<SetupPageProps> = (props) => {
  const [username, setUsername] = createSignal('')
  const [password, setPassword] = createSignal('')
  const [confirmPassword, setConfirmPassword] = createSignal('')
  const [loading, setLoading] = createSignal(false)
  const [error, setError] = createSignal('')
  const [showEnvCheckModal, setShowEnvCheckModal] = createSignal(false)

  const handleSubmit = async (e: Event) => {
    e.preventDefault()
    
    if (!username().trim() || !password().trim()) {
      setError('用户名和密码不能为空')
      return
    }
    
    if (password() !== confirmPassword()) {
      setError('密码确认不匹配')
      return
    }
    
    if (password().length < 6) {
      setError('密码长度至少6位')
      return
    }

    setLoading(true)
    setError('')

    try {
      const response = await fetch('/api/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: username().trim(),
          password: password()
        })
      })

      const data = await response.json()
      
      if (data.success) {
        props.onSetupComplete()
      } else {
        setError(data.message || '设置失败')
      }
    } catch (error) {
      setError('网络错误，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div class="min-h-screen bg-base-200 flex items-center justify-center">
      <div class="card w-96 bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-center justify-center mb-6">初始化设置</h2>
          <p class="text-center text-base-content/70 mb-4">
            首次使用需要设置管理员账号
          </p>
          
          <form onSubmit={handleSubmit}>
            <div class="form-control w-full mb-4">
              <label class="label">
                <span class="label-text">管理员用户名</span>
              </label>
              <input
                type="text"
                placeholder="请输入用户名"
                class="input input-bordered w-full"
                value={username()}
                onInput={(e) => setUsername(e.target.value)}
                disabled={loading()}
                required
              />
            </div>
            
            <div class="form-control w-full mb-4">
              <label class="label">
                <span class="label-text">密码</span>
              </label>
              <input
                type="password"
                placeholder="请输入密码"
                class="input input-bordered w-full"
                value={password()}
                onInput={(e) => setPassword(e.target.value)}
                disabled={loading()}
                required
              />
            </div>
            
            <div class="form-control w-full mb-6">
              <label class="label">
                <span class="label-text">确认密码</span>
              </label>
              <input
                type="password"
                placeholder="请再次输入密码"
                class="input input-bordered w-full"
                value={confirmPassword()}
                onInput={(e) => setConfirmPassword(e.target.value)}
                disabled={loading()}
                required
              />
            </div>
            
            {error() && (
              <div class="alert alert-error mb-4">
                <span>{error()}</span>
              </div>
            )}
            
            <div class="form-control">
              <button
                type="submit"
                class="btn btn-primary"
                disabled={loading()}
              >
                {loading() && <span class="loading loading-spinner"></span>}
                {loading() ? '设置中...' : '完成设置'}
              </button>
            </div>
          </form>
          
          {/* Environment Check Button */}
          <div class="divider">或</div>
          <button
            onClick={() => setShowEnvCheckModal(true)}
            class="btn btn-secondary btn-block"
            disabled={loading()}
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            环境检查
          </button>
        </div>
      </div>
      
      {/* Environment Check Modal */}
      <EnvironmentCheckModal
        isOpen={showEnvCheckModal()}
        onClose={() => setShowEnvCheckModal(false)}
      />
    </div>
  )
}

export default SetupPage