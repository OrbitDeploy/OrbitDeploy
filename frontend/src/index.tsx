import { render } from 'solid-js/web'
import { QueryClient, QueryClientProvider } from '@tanstack/solid-query' // 1. 引入依赖
import App from './App'
import './index.css'
import { installAuthFetchInterceptor } from './lib/apiClient'

// Ensure fetch interceptor is installed before any fetch/createFetch usage
try { installAuthFetchInterceptor() } catch {}

// 2. 创建 QueryClient 实例
const queryClient = new QueryClient();

// 3. 用 Provider 包裹 App 组件
render(
  () => (
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  ),
  document.getElementById('root') as HTMLElement
)