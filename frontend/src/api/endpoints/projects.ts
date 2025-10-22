// src/api/endpoints/projects.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const projectsEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/projects", "method": "GET" },
  "create": { "url": "/projects", "method": "POST" },
  "getByUid": { "url": "/projects/{projectId}", "method": "GET" },
  "getByName": { "url": "/projects/by-name/{name}", "method": "GET" },
  "createApp": { "url": "/projects/{projectId}/apps", "method": "POST" },
  "listApps": { "url": "/projects/{projectId}/apps", "method": "GET" },
  "listAppsByName": { "url": "/projects/by-name/{name}/apps", "method": "GET" },
  "getAppByName": { "url": "/projects/by-name/{projectName}/apps/by-name/{appName}", "method": "GET" }
};

registerEndpoints('projects', projectsEndpoints);

// --- 导出的函数 ---

export function listProjectsEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('projects', 'list');
}

export function createProjectEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('projects', 'create');
}

export function getProjectByUidEndpoint(projectId: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('projects', 'getByUid', { projectId });
}

export function getProjectByNameEndpoint(name: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('projects', 'getByName', { name });
}

export function createAppEndpoint(projectId: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('projects', 'createApp', { projectId });
}

export function listAppsEndpoint(projectId: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('projects', 'listApps', { projectId });
}

export function listAppsByNameEndpoint(name: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('projects', 'listAppsByName', { name });
}

export function getAppByNameEndpoint(projectName: string, appName: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('projects', 'getAppByName', { projectName, appName });
}
