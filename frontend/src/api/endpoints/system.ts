// src/api/endpoints/system.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const systemEndpoints: { [action: string]: EndpointConfig } = {
  "getSetting": { "url": "/system/settings/{key}", "method": "GET" },
  "updateSetting": { "url": "/system/settings/{key}", "method": "PUT" },
};

registerEndpoints('system', systemEndpoints);

// --- 导出的函数 ---

export function getSystemSettingEndpoint(key: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('system', 'getSetting', { key });
}

export function updateSystemSettingEndpoint(key: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('system', 'updateSetting', { key });
}
