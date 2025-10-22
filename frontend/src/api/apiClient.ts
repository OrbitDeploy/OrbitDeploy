// 这是一个通用的 API 请求客户端，负责执行 fetch

interface ApiMutateOptions {
  method: 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  body?: unknown;
  headers?: Record<string, string>;
}

// 统一的错误处理器
async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ message: response.statusText }));
    throw new Error(errorData.message || 'An unknown error occurred');
  }
  // 如果响应体可能为空 (例如 204 No Content), 进行处理
  if (response.status === 204 || response.headers.get('Content-Length') === '0') {
      return null as T;
  }
  
  const json = await response.json();

  // Check for the wrapper
  if (json && typeof json === 'object' && 'success' in json) {
    if (json.success) {
      return json.data as T;
    } else {
      throw new Error(json.message || 'An API error occurred');
    }
  }

  // If no wrapper, return the whole object
  return json as T;
}


/**
 * 用于发起 GET 请求
 */
export async function apiGet<T>(url: string): Promise<T> {
  const response = await fetch(url);
  return handleResponse<T>(response);
}

/**
 * 用于发起数据变更请求 (POST, PUT, DELETE, etc.)
 * 这个函数现在会严格遵守传入的 method 参数
 */
export async function apiMutate<T>(url: string, options: ApiMutateOptions): Promise<T> {
  const { method, body, headers = {} } = options;

  const config: RequestInit = {
    method: method,
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
  };

  if (body !== undefined) {
    config.body = JSON.stringify(body);
  }

  const response = await fetch(url, config);
  return handleResponse<T>(response);
}
