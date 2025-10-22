API 端点迁移指南：从 JSON 到模块化函数本文档提供了一个清晰的操作流程，旨在指导开发者如何将旧的、集中在 JSON 文件中的 API 定义，迁移到新的、模块化、类型安全的 TypeScript 文件中。我们假设核心文件 _core.ts 及其中的 registerEndpoints, getApiEndpoint, EndpointConfig 等功能已经实现。第 1 步：分析源 JSON 结构在迁移开始前，首先从旧的 api.config.json 文件中找到你想要迁移的模块部分。示例：假设我们要迁移一个新的模块 "users"。// frontend\src\api\config.ts (旧)
...
  "users": {
    "list": { "url": "/users", "method": "GET" },
    "getById": { "url": "/users/{id}", "method": "GET" },
    "create": { "url": "/users", "method": "POST" },
    "updateProfile": { "url": "/users/{id}/profile", "method": "PUT" }
  }
...
从这段 JSON 中，我们需要提取以下关键信息：模块名 (Module Name)："users"。动作名 (Action Name)："list", "getById", "create", "updateProfile"。端点配置 (Endpoint Config)：每个动作对应的 { "url": "...", "method": "..." } 对象。URL 参数 (URL Parameters)：URL 字符串中用花括号 {} 包裹的部分，例如 {id}。第 2 步：创建模块文件并构造函数现在，我们将为 "users" 模块创建一个新的 API 文件 src/api/endpoints/users.ts。2.1 定义端点并注册创建文件 users.ts。将第 1 步中分析的 JSON 内容复制过来，并定义为一个带有显式类型的 TypeScript 常量。在常量定义下方，立即调用 registerEndpoints 函数进行注册。// src/api/endpoints/users.ts

// 从核心文件导入所需的函数和类型
import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

// 1. 定义端点常量，并为其添加显式类型
const userEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/users", "method": "GET" },
  "getById": { "url": "/users/{id}", "method": "GET" },
  "create": { "url": "/users", "method": "POST" },
  "updateProfile": { "url": "/users/{id}/profile", "method": "PUT" }
};

// 2. 立即调用注册函数
registerEndpoints('users', userEndpoints);

// ... 下一步将在此处添加导出的函数 ...
重点: 必须添加 : { [action: string]: EndpointConfig } 这个类型注解。它能确保 TypeScript 将 "GET", "POST" 等识别为 ApiMethod 类型，而不是普通的 string，从而避免类型错误。2.2 逐一构造导出函数接下来，为 userEndpoints 中的每一个动作 (Action) 创建一个对应的导出函数。遵循以下模式：示例 A：list 动作 (无参数)分析: JSON 行为 "list": { "url": "/users", "method": "GET" }。URL 中没有参数。命名: 函数应命名为 listUsersEndpoint。参数: 函数不需要接收参数。实现: 调用 getApiEndpoint('users', 'list')。返回类型: method 是 GET，所以返回 ApiEndpoint<'GET'>。最终代码:export function listUsersEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('users', 'list');
}
示例 B：getById 动作 (有参数)分析: JSON 行为 "getById": { "url": "/users/{id}", "method": "GET" }。URL 中有一个参数 {id}。命名: 函数应命名为 getUserByIdEndpoint。参数: 函数需要接收一个 id: string (或 number) 参数。实现: 调用 getApiEndpoint 时，将参数作为第三个参数传入：getApiEndpoint('users', 'getById', { id })。返回类型: method 是 GET，所以返回 ApiEndpoint<'GET'>。最终代码:export function getUserByIdEndpoint(id: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('users', 'getById', { id });
}
按照这个模式，为 userEndpoints 对象中的所有条目完成函数的创建。最终的 users.ts 文件// src/api/endpoints/users.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const userEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/users", "method": "GET" },
  "getById": { "url": "/users/{id}", "method": "GET" },
  "create": { "url": "/users", "method": "POST" },
  "updateProfile": { "url": "/users/{id}/profile", "method": "PUT" }
};

registerEndpoints('users', userEndpoints);

// --- 导出的函数 ---

export function listUsersEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('users', 'list');
}

export function getUserByIdEndpoint(id: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('users', 'getById', { id });
}

export function createUserEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('users', 'create');
}

export function updateUserProfileEndpoint(id: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('users', 'updateProfile', { id });
}
