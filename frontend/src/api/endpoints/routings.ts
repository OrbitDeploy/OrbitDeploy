// src/api/endpoints/routings.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const routingsEndpoints: { [action: string]: EndpointConfig } = {
  "create": { "url": "/apps/{uid}/routings", "method": "POST" },
  "list": { "url": "/apps/{uid}/routings", "method": "GET" },
  "update": { "url": "/routings/{routingId}", "method": "PUT" },
  "delete": { "url": "/routings/{routingId}", "method": "DELETE" }
};

registerEndpoints('routings', routingsEndpoints);

// --- 导出的函数 ---

export function createRoutingEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('routings', 'create', { uid });
}

export function listRoutingsEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('routings', 'list', { uid });
}

export function updateRoutingEndpoint(routingId: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('routings', 'update', { routingId });
}

export function deleteRoutingEndpoint(routingId: string): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('routings', 'delete', { routingId });
}
