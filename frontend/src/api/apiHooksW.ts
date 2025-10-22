import type { Accessor } from 'solid-js'

import {
  useQuery,
  useMutation,
  UseMutationOptions ,
  type QueryKey,
  type SolidQueryOptions,
  // Import the specific result types for our overloads
  type UseQueryResult,
  type DefinedUseQueryResult,
} from '@tanstack/solid-query';
import { createMemo } from 'solid-js';
import { apiGet, apiMutate } from '../api/apiClient';
import type { ApiEndpoint, ApiMutationMethod } from '../api/endpoints/_core.ts';

type ApiQueryOptions<T> = Omit<SolidQueryOptions<T>, 'queryKey' | 'queryFn'>;


export function useApiQuery<T>(
  queryKey: QueryKey | (() => QueryKey),
  urlFn: () => string | null,
  options: ApiQueryOptions<T> & { initialData: T | (() => T) }
): DefinedUseQueryResult<T, Error>;

export function useApiQuery<T>(
  queryKey: QueryKey | (() => QueryKey),
  urlFn: () => string | null,
  options?: Omit<ApiQueryOptions<T>, 'initialData'>
): UseQueryResult<T, Error>;

export function useApiQuery<T>(
  queryKey: QueryKey | (() => QueryKey),
  urlFn: () => string | null,
  options?: ApiQueryOptions<T>
) {
  const resolvedKey = createMemo(() => (typeof queryKey === 'function' ? queryKey() : queryKey));


  return useQuery(() => ({
    queryKey: resolvedKey(),
    queryFn: () => {
      const url = urlFn();
      if (url === null) {
        return Promise.reject(new Error('URL is null'));
      }
      return apiGet<T>(url);
    },
    enabled: urlFn() != null,
    ...options,
  }) as any);
}


// 解包 Accessor 内部类型的工具类型
type UnwrapAccessor<T> = T extends Accessor<infer U> ? U : T

// 重新定义 ApiMutationOptions
type ApiMutationOptions<TData, TVariables> =
  Omit<UnwrapAccessor<UseMutationOptions<TData, Error, TVariables>>, 'mutationFn'> & {
    body?: (variables: TVariables) => unknown
  }

export function useApiMutation<TData = unknown, TVariables = void>(
  endpointOrFn: ApiEndpoint<ApiMutationMethod> | ((variables: TVariables) => ApiEndpoint<ApiMutationMethod>),
  options?: ApiMutationOptions<TData, TVariables>
) {
  const { body: bodyFn, ...restOptions } = options || {};

  return useMutation<TData, Error, TVariables>(() => ({
    mutationFn: async (variables: TVariables) => {
      const endpoint = typeof endpointOrFn === 'function' 
        ? endpointOrFn(variables) 
        : endpointOrFn;
      
      const body = bodyFn ? bodyFn(variables) : variables;
      
      return apiMutate<TData>(endpoint.url, {
        method: endpoint.method,
        body,
      });
    },
    ...restOptions,
  }));
}