// src/api/request.ts
import axios from 'axios';

// 创建axios实例，不设置baseURL，直接使用相对路径
// 这样可以利用Vite的代理配置在开发环境中工作
const apiClient = axios.create({
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 可以在这里添加请求和响应拦截器，用于统一处理token、错误等
apiClient.interceptors.request.use(config => {
  // 统一逻辑，例如添加认证头
  return config;
});

apiClient.interceptors.response.use(
  response => response,
  error => {
    // 统一错误处理
    console.error('API Request Error:', error);
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    const errorMessage = error && typeof error === 'object' && 'message' in error 
      ? String(error.message) 
      : 'API request failed';
    return Promise.reject(new Error(errorMessage));
  }
);

export default apiClient;