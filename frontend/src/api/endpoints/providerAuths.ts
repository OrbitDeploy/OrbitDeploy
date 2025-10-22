// src/api/endpoints/providerAuths.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const providerAuthsEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/provider-auths", "method": "GET" },
  "create": { "url": "/provider-auths", "method": "POST" },
  "getById": { "url": "/provider-auths/{uid}", "method": "GET" },
  "update": { "url": "/provider-auths/{uid}", "method": "PUT" },
  "delete": { "url": "/provider-auths/{uid}", "method": "DELETE" },
  "activate": { "url": "/provider-auths/{uid}/activate", "method": "POST" },
  "deactivate": { "url": "/provider-auths/{uid}/deactivate", "method": "POST" },
  "repositories": { "url": "/provider-auths/{uid}/repositories", "method": "GET" },
  "branches": { "url": "/provider-auths/{uid}/repositories/branches", "method": "GET" },
  "githubAppManifest": { "url": "/providers/github/app-manifest", "method": "GET" },
  "githubAppCallback": { "url": "/providers/github/app-callback", "method": "POST" },
  "githubInstall": { "url": "/provider-auths/{uid}/github-install", "method": "POST" }
};

registerEndpoints('providerAuths', providerAuthsEndpoints);

// --- 导出的函数 ---

export function listProviderAuthsEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('providerAuths', 'list');
}

export function createProviderAuthEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('providerAuths', 'create');
}

export function getProviderAuthByIdEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('providerAuths', 'getById', { uid });
}

export function updateProviderAuthEndpoint(uid: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('providerAuths', 'update', { uid });
}

export function deleteProviderAuthEndpoint(uid: string): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('providerAuths', 'delete', { uid });
}

export function activateProviderAuthEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('providerAuths', 'activate', { uid });
}

export function deactivateProviderAuthEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('providerAuths', 'deactivate', { uid });
}

export function getProviderAuthRepositoriesEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('providerAuths', 'repositories', { uid });
}

export function getProviderAuthBranchesEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('providerAuths', 'branches', { uid });
}

export function getGithubAppManifestEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('providerAuths', 'githubAppManifest');
}

export function githubAppCallbackEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('providerAuths', 'githubAppCallback');
}

export function githubInstallEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('providerAuths', 'githubInstall', { uid });
}
