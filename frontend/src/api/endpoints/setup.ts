// src/api/endpoints/setup.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const setupEndpoints: { [action: string]: EndpointConfig } = {
  "check": { "url": "/setup/check", "method": "GET" }
};

registerEndpoints('setup', setupEndpoints);

// --- 导出的函数 ---

export function checkSetupEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('setup', 'check');
}