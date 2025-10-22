// src/api/endpoints/environments.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const environmentsEndpoints: { [action: string]: EndpointConfig } = {
  "check": { "url": "/environment/check", "method": "GET" },
  "installPodman": { "url": "/environment/install-podman", "method": "POST" },
  "installCaddy": { "url": "/environment/install-caddy", "method": "POST" }
};

registerEndpoints('environments', environmentsEndpoints);

// --- 导出的函数 ---

export function checkEnvironmentEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('environments', 'check');
}

export function installPodmanEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('environments', 'installPodman');
}

export function installCaddyEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('environments', 'installCaddy');
}