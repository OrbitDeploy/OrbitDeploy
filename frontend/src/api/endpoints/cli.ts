// src/api/endpoints/cli.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const cliEndpoints: { [action: string]: EndpointConfig } = {
  "deviceAuthSessions": { "url": "/cli/device-auth/sessions/{sessionId}", "method": "GET" },
  "deviceAuthConfirm": { "url": "/cli/device-auth/confirm", "method": "POST" }
};

registerEndpoints('cli', cliEndpoints);

// --- 导出的函数 ---

export function deviceAuthSessionsEndpoint(sessionId: string): ApiEndpoint<'GET'> {
  return getApiEndpoint('cli', 'deviceAuthSessions', { sessionId });
}

export function deviceAuthConfirmEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('cli', 'deviceAuthConfirm');
}