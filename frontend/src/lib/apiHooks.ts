import {
  useQuery,
  useMutation,
  type UseQueryOptions,
  type QueryKey,
} from '@tanstack/solid-query'
import { createMemo } from 'solid-js'
import { apiGet, apiMutate } from './apiClient'

/**
 * 一个封装了 apiGet 的通用 useApiQuery Hook。
 */
export function useApiQuery<T>(
  queryKey: QueryKey | (() => QueryKey),
  url: string | (() => string | null),
  options?: Omit<UseQueryOptions<T>, 'queryKey' | 'queryFn'> & { enabled?: boolean | (() => boolean) },
) {
  // Create memos for reactive values to ensure proper reactivity
  const resolvedUrl = createMemo(() => typeof url === 'function' ? (url as () => string | null)() : url)
  const resolvedKey = createMemo(() => typeof queryKey === 'function' ? (queryKey as () => QueryKey)() : queryKey)
  const { enabled: rawEnabled, ...rest } = (options || {}) as any
  const resolvedEnabled = createMemo(() => {
    const enabled = typeof rawEnabled === 'function' ? !!rawEnabled() : rawEnabled
    return enabled ?? (resolvedUrl() != null)
  })

  return useQuery<T>(() => ({
    queryKey: resolvedKey(),
    // 注意：当 resolvedUrl 为 null 时，不应调用 queryFn（需结合 enabled 控制）
    queryFn: () => {
      return apiGet<T>(resolvedUrl() as string)
    },
    enabled: resolvedEnabled(),
    ...rest,
  }))
}




export function useApiMutation<TData = unknown, TVariables = void>(
  // 参数1现在可以是：URL字符串、URL构造函数、或完整的 mutation 函数
  mutationFnOrUrl: string | ((variables: TVariables) => string) | ((variables: TVariables) => Promise<TData>),
  options?: any,
): ReturnType<typeof useMutation<TData, Error, TVariables>> {

  // 如果传入的是一个完整的、返回 Promise 的 mutation 函数，则直接使用 (保持了旧的灵活性)
  // (这是一个简单的判断，实际项目中可能需要更严谨的方式来区分函数类型)
  if (typeof mutationFnOrUrl === 'function' && mutationFnOrUrl.toString().includes('Promise')) {
      return useMutation<TData, Error, TVariables>(() => ({
        // FIX 1: Assert the type here to ensure it matches what useMutation expects.
        mutationFn: mutationFnOrUrl as (variables: TVariables) => Promise<TData>,
        ...options,
      }))
  }

  // ==================== 升级的核心逻辑 ====================
  // 如果传入的是 URL 字符串 或 URL 构造函数
  const { method = 'POST', headers, defaultErrorMsg, ...rest } = options || {}

  // FIX 2: Assert that the remaining possibilities are what we expect.
  // This is the key to solving the error.
  const urlOrUrlFn = mutationFnOrUrl as string | ((variables: TVariables) => string);

  return useMutation<TData, Error, TVariables>(() => ({
    mutationFn: (variables: TVariables) => {
      // 动态判断 URL：现在 TypeScript knows `urlOrUrlFn`不能返回 Promise.
      const finalUrl = typeof urlOrUrlFn === 'function'
        ? urlOrUrlFn(variables) // 这现在正确解析为 `string`
        : urlOrUrlFn;           // 这也是一个 `string`

      const body = options?.body ? options.body(variables) : variables;

      // 错误现在消失了，因为 `finalUrl` 被保证是一个 `string`.
      return apiMutate<TData>(finalUrl, { method, body, headers, defaultErrorMsg })
    },
    ...rest,
  }))
}