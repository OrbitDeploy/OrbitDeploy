// src/api/endpoints/images.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const imagesEndpoints: { [action: string]: EndpointConfig } = {
  "buildFromGitHub": { "url": "/images/build-from-github", "method": "POST" }
};

registerEndpoints('images', imagesEndpoints);

// --- 导出的函数 ---

export function buildFromGitHubEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('images', 'buildFromGitHub');
}