/**
 * 通用 API 客户端封装 - 支持 JWT 认证
 * 目标：
 * - 统一解析后端的标准响应：{ success: boolean, data?: T, message?: string }
 * - 对照 doc/FRONTEND_API_RESPONSE_ADAPTATION.md、doc/createFetch.md 和 ExamplesPage 的用法
 * - 提供 GET 读取类和变更类（POST/PUT/DELETE）便捷函数
 * - 支持 JWT 访问令牌自动附加和刷新
 * - 中文注释，先供试用，后续按需在各页面替换使用
 */

import { getAuthApiUrl } from '../api/config'

export interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
}

export class ApiError extends Error {
  status?: number
  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// 全局访问令牌管理
let globalAccessToken: string | null = null
let isRefreshing = false
let refreshPromise: Promise<string | null> | null = null

// 保留原始 fetch 的引用，用于在拦截器中避免递归
let originalFetch: typeof fetch | null = null

/**
 * 设置全局访问令牌
 */
export function setGlobalAccessToken(token: string | null) {
  globalAccessToken = token
}

/**
 * 获取全局访问令牌
 */
export function getGlobalAccessToken(): string | null {
  return globalAccessToken
}

/**
 * 刷新访问令牌
 */
async function refreshAccessToken(): Promise<string | null> {
  // 如果已经在刷新中，返回现有的 Promise
  if (isRefreshing && refreshPromise) {
    return refreshPromise
  }

  isRefreshing = true
  refreshPromise = (async () => {
    try {
      const response = await fetch(getAuthApiUrl('refreshToken'), {
        method: 'POST',
        credentials: 'include', // 包含 refresh token cookie
      })

      if (response.ok) {
        const data = await response.json()
        if (data.success && data.data.access_token) {
          globalAccessToken = data.data.access_token
          return data.data.access_token
        }
      }
      
      // 刷新失败，清空令牌
      globalAccessToken = null
      return null
    } catch (error) {
      console.error('Token refresh failed:', error)
      globalAccessToken = null
      return null
    } finally {
      isRefreshing = false
      refreshPromise = null
    }
  })()

  return refreshPromise
}

/**
 * 安装全局 fetch 拦截器：
 * - 自动附加 Authorization: Bearer <access_token>
 * - 401 时自动尝试刷新并重试
 */
export function installAuthFetchInterceptor() {
  if (typeof window === 'undefined') return
  if (!originalFetch) {
    originalFetch = window.fetch.bind(window)
  }
  // 避免重复安装
  const alreadyWrapped = (window.fetch as any).__isAuthWrapped
  if (alreadyWrapped) return

  // 暴露一个全局方法便于 SSE/WS 取 token（只读）
  ;(window as any).getAccessToken = getGlobalAccessToken

  const wrappedFetch = async (url: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const urlStr = typeof url === 'string' ? url : url.toString()
    const isRefreshEndpoint = urlStr.includes(getAuthApiUrl('refreshToken'))

    // 准备请求头
    const headers = new Headers(init?.headers)

    // 如果有访问令牌，添加到请求头（刷新接口除外）
    if (globalAccessToken && !isRefreshEndpoint) {
      headers.set('Authorization', `Bearer ${globalAccessToken}`)
    }

    // 第一次请求
    let response = await (originalFetch || fetch)(url, { ...init, headers })

    // 如果返回 401，尝试刷新访问令牌
    if (response.status === 401 && !isRefreshEndpoint) {
      const newToken = await refreshAccessToken()
      if (newToken) {
        headers.set('Authorization', `Bearer ${newToken}`)
        response = await (originalFetch || fetch)(url, { ...init, headers })
      }
    }

    return response
  }
  ;(wrappedFetch as any).__isAuthWrapped = true
  window.fetch = wrappedFetch as typeof fetch
}

// 尝试在模块加载时安装拦截器（多次调用安全）
if (typeof window !== 'undefined') {
  try { installAuthFetchInterceptor() } catch {}
}

/**
 * 带 JWT 认证的 fetch 函数（仍保留便于直接调用）
 */
async function authenticatedFetch(url: string, init?: RequestInit): Promise<Response> {
  // 确保全局拦截器已安装
  installAuthFetchInterceptor()
  // 直接使用全局 fetch（已拦截）
  return fetch(url, init)
}

/**
 * 统一解析响应体：
 * - 支持后端标准响应：{ success: boolean, data?: T, message?: string }
 * - 也支持直接数据响应（无包装）：直接返回 JSON
 * - 始终尝试解析 JSON；无法解析时按失败处理
 * - success !== true 一律抛错，错误信息取 message 字段，兜底为通用文案
 */
export async function parseJsonOrThrow<T>(response: Response, defaultErrorMsg = '操作失败'): Promise<T> {
  let json: any = null
  try {
    json = await response.json()
  } catch {
    // 无法解析 JSON，也视为失败
    throw new ApiError(defaultErrorMsg, response.status)
  }

  // 检查是否是标准包装响应
  if (json && typeof json === 'object' && 'success' in json) {
    if (json.success !== true) {
      const msg = json.message || defaultErrorMsg
      throw new ApiError(msg, response.status)
    }
    return json.data as T
  }

  // 如果不是包装响应，直接返回 JSON（适用于直接数据 API）
  return json as T
}

/**
 * 读取类 GET：返回 data 部分，自动处理 JWT 认证
 * 用法：const data = await apiGet<MyType>('/api/foo')
 */
export async function apiGet<T>(url: string, init?: RequestInit & { defaultErrorMsg?: string }): Promise<T> {
  const { defaultErrorMsg, ...rest } = init || {}
  const resp = await authenticatedFetch(url, { method: 'GET', ...rest })
  return parseJsonOrThrow<T>(resp, defaultErrorMsg)
}

/**
 * 变更类请求：POST/PUT/DELETE，自动处理 JWT 认证
 * - 自动设置 Content-Type: application/json（可被 init.headers 覆盖）
 * - body 可传对象，内部会 JSON.stringify
 * - 返回 data 部分
 */
export async function apiMutate<T>(
  url: string,
  options: {
    method?: 'POST' | 'PUT' | 'DELETE' | 'PATCH'
    body?: unknown
    headers?: Record<string, string>
    defaultErrorMsg?: string
  } = {}
): Promise<T> {
  const { method = 'POST', body, headers, defaultErrorMsg } = options
  const init: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json', ...(headers || {}) },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  }
  const resp = await authenticatedFetch(url, init)
  return parseJsonOrThrow<T>(resp, defaultErrorMsg)
}

/**
 * 工具：安全获取 data（用于需要容错的地方）。
 * - 若失败则返回 fallback，不抛异常。
 * - 适用于页面初次渲染时不希望中断 UI 的场景。
 */
export async function tryApiGet<T>(url: string, fallback: T, init?: RequestInit & { defaultErrorMsg?: string }): Promise<T> {
  try {
    return await apiGet<T>(url, init)
  } catch {
    return fallback
  }
}

/** Raw mutate that returns status and json without throwing */
export async function apiMutateRaw(
  url: string,
  options: {
    method?: 'POST' | 'PUT' | 'DELETE' | 'PATCH'
    body?: unknown
    headers?: Record<string, string>
  } = {}
): Promise<{ status: number; json: any }> {
  const { method = 'POST', body, headers } = options
  const init: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json', ...(headers || {}) },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  }
  const resp = await authenticatedFetch(url, init)
  let json: any = null
  try { json = await resp.json() } catch {}
  return { status: resp.status, json }
}

/**
 * 示例：从 ExamplesPage 迁移来的调用风格
 * - 创建：await apiMutate('/api/examples', { method: 'POST', body: { name } })
 * - 更新：await apiMutate(`/api/examples/${id}`, { method: 'PUT', body: { name } })
 * - 删除：await apiMutate(`/api/examples/${id}`, { method: 'DELETE' })
 * - 查询：const list = await apiGet<Example[]>('/api/examples')
 */
