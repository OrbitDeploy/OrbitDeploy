// src/api/endpoints/_core.ts

// --- 类型定义 ---
export type ApiMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
export type ApiMutationMethod = Exclude<ApiMethod, 'GET'>;
export interface ApiEndpoint<TMethod extends ApiMethod = ApiMethod> {
  url: string;
  method: TMethod;
}
export interface ApiResponse<T> {
  data: T;
  success?: boolean;
  message?: string;
  code?: number;
}
// 将 EndpointConfig 导出，以便其他模块可以使用它
export type EndpointConfig = { url: string; method: ApiMethod };

// --- 内存中的端点配置存储 ---
const apiConfigStore: { [module: string]: { [action: string]: EndpointConfig } } = {};


// --- 用于注册模块端点的导出函数 ---
export function registerEndpoints(moduleName:string, endpoints: { [action: string]: EndpointConfig }) {
    if (apiConfigStore[moduleName]) {
        console.warn(`API module "${moduleName}" is being registered more than once.`);
    }
    apiConfigStore[moduleName] = endpoints;
}


// --- 内部私有辅助函数 ---
const LEGACY_API_PREFIX = '/api';

function replaceUrlParams(template: string, params: Record<string, string | number>): string {
  let url = template;
  for (const [key, value] of Object.entries(params)) {
    url = url.replace(`{${key}}`, String(value));
  }
  return url;
}

function buildApiUrl(endpoint: string): string {
  const normalizedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
  return `${LEGACY_API_PREFIX}${normalizedEndpoint}`;
}


// --- 内部函数 (用于读取内存存储) ---
function getEndpointConfig(module: string, action: string): EndpointConfig {
  const endpointConfig = apiConfigStore[module]?.[action];

  if (!endpointConfig || typeof endpointConfig.url !== 'string' || typeof endpointConfig.method !== 'string') {
    throw new Error(`API endpoint config for "${module}.${action}" is not registered or is invalid.`);
  }

  return endpointConfig;
}


// --- 核心导出函数 ---
export function getApiEndpoint<TMethod extends ApiMethod>(
  module: string,
  action: string,
  params?: Record<string, string | number>
): ApiEndpoint<TMethod> {
  const { url: urlTemplate, method } = getEndpointConfig(module, action);
  const populatedUrl = params ? replaceUrlParams(urlTemplate, params) : urlTemplate;

  return {
    url: buildApiUrl(populatedUrl),
    method: method as TMethod,
  };
}