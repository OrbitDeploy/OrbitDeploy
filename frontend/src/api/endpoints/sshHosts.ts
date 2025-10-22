// src/api/endpoints/sshHosts.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const sshHostsEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/ssh-hosts", "method": "GET" },
  "create": { "url": "/ssh-hosts", "method": "POST" },
  "getById": { "url": "/ssh-hosts/{uid}", "method": "GET" },
  "update": { "url": "/ssh-hosts/{uid}", "method": "PUT" },
  "delete": { "url": "/ssh-hosts/{uid}", "method": "DELETE" },
  "test": { "url": "/ssh-hosts/{uid}/test", "method": "POST" }
};

registerEndpoints('sshHosts', sshHostsEndpoints);

// --- 导出的函数 ---

export function listSshHostsEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('sshHosts', 'list');
}

export function createSshHostEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('sshHosts', 'create');
}

export function getSshHostByIdEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('sshHosts', 'getById', { uid });
}

export function updateSshHostEndpoint(uid: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('sshHosts', 'update', { uid });
}

export function deleteSshHostEndpoint(uid: string): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('sshHosts', 'delete', { uid });
}

export function testSshHostEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('sshHosts', 'test', { uid });
}
