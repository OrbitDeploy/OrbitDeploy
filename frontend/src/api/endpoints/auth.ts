// src/api/endpoints/auth.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const authEndpoints: { [action: string]: EndpointConfig } = {
  "login": { "url": "/auth/login", "method": "POST" },
  "logout": { "url": "/auth/logout", "method": "POST" },
  "refreshToken": { "url": "/auth/refresh_token", "method": "POST" },
  "status": { "url": "/auth/status", "method": "GET" },
  "changePassword": { "url": "/auth/change-password", "method": "PUT" },
  "setup2FA": { "url": "/2fa/setup", "method": "POST" },
  "verify2FA": { "url": "/2fa/verify", "method": "POST" },
  "login2FA": { "url": "/2fa/login", "method": "POST" },
  "disable2FA": { "url": "/2fa/disable", "method": "DELETE" }
};

registerEndpoints('auth', authEndpoints);

// --- 导出的函数 ---

export function loginEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('auth', 'login');
}

export function logoutEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('auth', 'logout');
}

export function refreshTokenEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('auth', 'refreshToken');
}

export function statusEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('auth', 'status');
}

export function changePasswordEndpoint(): ApiEndpoint<'PUT'> {
  return getApiEndpoint('auth', 'changePassword');
}

export function setup2FAEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('auth', 'setup2FA');
}

export function verify2FAEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('auth', 'verify2FA');
}

export function login2FAEndpoint(): ApiEndpoint<'POST'> {
  return getApiEndpoint('auth', 'login2FA');
}

export function disable2FAEndpoint(): ApiEndpoint<'DELETE'> {
  return getApiEndpoint('auth', 'disable2FA');
}