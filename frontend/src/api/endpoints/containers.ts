// src/api/endpoints/containers.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const containersEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/containers", "method": "GET" },
  "checkEnv": { "url": "/containers/check-env", "method": "GET" }
};

registerEndpoints('containers', containersEndpoints);

// --- 导出的函数 ---

export function listContainersEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('containers', 'list');
}

export function checkEnvEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('containers', 'checkEnv');
}