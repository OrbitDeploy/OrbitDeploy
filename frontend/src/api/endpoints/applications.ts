// src/api/endpoints/applications.ts

// 导入所需的函数和类型, 新增 EndpointConfig
import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

// 1. 为 applicationEndpoints 常量添加显式类型
const applicationEndpoints: { [action: string]: EndpointConfig } = {
  "create": { "url": "/apps", "method": "POST" },
  "getById": { "url": "/apps/{uid}", "method": "GET" },
  "getByName": { "url": "/apps/by-name/{name}", "method": "GET" },
  "update": { "url": "/apps/{uid}", "method": "PUT" },
  "delete": { "url": "/apps/{uid}", "method": "DELETE" },
  "status": { "url": "/apps/{uid}/status", "method": "GET" },
  "logs": { "url": "/apps/{uid}/logs", "method": "GET" },
  "deployments": { "url": "/apps/{uid}/deployments", "method": "GET" },
  "createDeployment": { "url": "/apps/{uid}/deployments", "method": "POST" },
  "runningDeployments": { "url": "/apps/{identifier}/deployments/running", "method": "GET" },
  "releases": { "url": "/apps/{uid}/releases", "method": "GET" },
  "latestRelease": { "url": "/apps/{uid}/releases/latest", "method": "GET" },
  "configurations": { "url": "/apps/{uid}/configurations", "method": "GET" },
  "routings": { "url": "/apps/{uid}/routings", "method": "GET" },
  "tokens": { "url": "/apps/{uid}/tokens", "method": "GET" },
  "tokenCreate": { "url": "/apps/{uid}/tokens", "method": "POST" },
  "tokenUpdate": { "url": "/apps/{uid}/tokens/{tokenId}", "method": "PUT" },
  "tokenDelete": { "url": "/apps/{uid}/tokens/{tokenId}", "method": "DELETE" },
  "environmentVariables": { "url": "/apps/{uid}/environment-variables", "method": "GET" },
  "environmentVariableCreate": { "url": "/apps/{uid}/environment-variables", "method": "POST" },
  "environmentVariableUpdate": { "url": "/environment-variables/{envVarId}", "method": "PUT" },
  "environmentVariableDelete": { "url": "/environment-variables/{envVarId}", "method": "DELETE" }
};

// 2. 立即调用注册函数
registerEndpoints('applications', applicationEndpoints);


// 3. 导出所有具体的端点函数 (这部分代码无需修改)
export function createApplicationEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('applications', 'create');
}

export function getApplicationByIdEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'getById', { uid });
}

export function updateApplicationEndpoint(uid: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('applications', 'update', { uid });
}

export function deleteApplicationEndpoint(uid: string): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('applications', 'delete', { uid });
}

export function getApplicationByNameEndpoint(name: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'getByName', { name });
}

export function getApplicationStatusEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'status', { uid });
}

export function getApplicationLogsEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'logs', { uid });
}

export function getApplicationDeploymentsEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'deployments', { uid });
}

export function createApplicationDeploymentEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('applications', 'createDeployment', { uid });
}

export function getRunningDeploymentsEndpoint(identifier: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'runningDeployments', { identifier });
}

export function getApplicationReleasesEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'releases', { uid });
}

export function getLatestApplicationReleaseEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'latestRelease', { uid });
}

export function getApplicationConfigurationsEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'configurations', { uid });
}

export function getApplicationRoutingsEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'routings', { uid });
}

export function getApplicationTokensEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'tokens', { uid });
}

export function createApplicationTokenEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('applications', 'tokenCreate', { uid });
}

export function updateApplicationTokenEndpoint(uid: string, tokenId: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('applications', 'tokenUpdate', { uid, tokenId });
}

export function deleteApplicationTokenEndpoint(uid: string, tokenId: string): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('applications', 'tokenDelete', { uid, tokenId });
}

export function getEnvironmentVariablesEndpoint(uid: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('applications', 'environmentVariables', { uid });
}

export function createEnvironmentVariableEndpoint(uid: string): ApiEndpoint<'POST'> {
  return getApiEndpoint('applications', 'environmentVariableCreate', { uid });
}

export function updateEnvironmentVariableEndpoint(envVarId: string): ApiEndpoint<'PUT'> {
  return getApiEndpoint('applications', 'environmentVariableUpdate', { envVarId });
}

export function deleteEnvironmentVariableEndpoint(envVarId: string): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('applications', 'environmentVariableDelete', { envVarId });
}