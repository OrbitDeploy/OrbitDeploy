// src/api/endpoints/githubTokens.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const githubTokensEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/github-tokens", "method": "GET" },
  "create": { "url": "/github-tokens", "method": "POST" },
  "getById": { "url": "/github-tokens/{uid}", "method": "GET" },
  "update": { "url": "/github-tokens/{uid}", "method": "PUT" },
  "delete": { "url": "/github-tokens/{uid}", "method": "DELETE" },
  "test": { "url": "/github-tokens/{uid}/test", "method": "POST" }
};

registerEndpoints('githubTokens', githubTokensEndpoints);

// --- 导出的函数 ---

export function listGithubTokensEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('githubTokens', 'list');
}

export function createGithubTokenEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('githubTokens', 'create');
}

export function getGithubTokenByIdEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('githubTokens', 'getById', { uid });
}

export function updateGithubTokenEndpoint(uid: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('githubTokens', 'update', { uid });
}

export function deleteGithubTokenEndpoint(uid: string): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('githubTokens', 'delete', { uid });
}

export function testGithubTokenEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('githubTokens', 'test', { uid });
}