// src/api/endpoints/users.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const userEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/users", "method": "GET" },
  "getById": { "url": "/users/{uid}", "method": "GET" }
};

registerEndpoints('users', userEndpoints);

// --- 导出的函数 ---

export function listUsersEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('users', 'list');
}

export function getUserByIdEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('users', 'getById', { uid });
}