// src/api/endpoints/deployments.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const deploymentsEndpoints: { [action: string]: EndpointConfig } = {
  "getById": { "url": "/deployments/{uid}", "method": "GET" },
  "logs": { "url": "/deployments/{uid}/logs", "method": "GET" },
  "logsData": { "url": "/deployments/{uid}/logs-data", "method": "GET" },
  "restart": { "url": "/deployments/{uid}/restart", "method": "POST" },
  "status": { "url": "/deployments/{uid}/status", "method": "GET" }
};

registerEndpoints('deployments', deploymentsEndpoints);

// --- 导出的函数 ---

export function getDeploymentByIdEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('deployments', 'getById', { uid });
}

export function getDeploymentLogsEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('deployments', 'logs', { uid });
}

export function getDeploymentLogsDataEndpoint(uid: string, params?: { limit?: number; before_timestamp?: string }): ApiEndpoint<'GET'> {
  const endpoint = getApiEndpoint('deployments', 'logsData', { uid });
  if (params) {
    const queryParams = new URLSearchParams();
    if (params.limit) queryParams.append('limit', params.limit.toString());
    if (params.before_timestamp) queryParams.append('before_timestamp', params.before_timestamp);
    endpoint.url = `${endpoint.url}?${queryParams.toString()}`;
  }
  return endpoint;
}

export function restartDeploymentEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('deployments', 'restart', { uid });
}

export function getDeploymentStatusEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('deployments', 'status', { uid });
}
