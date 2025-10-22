// [面向未来] API版本前缀 (当前未使用，为未来扩展保留)
const API_PREFIX_V1 = '/api/v1';

// 在Go内嵌模式下，生产环境的BaseURL通常就是空字符串，因为是相对路径
// 开发环境下，我们指向Go服务的地址
// eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
const BASE_URL = import.meta.env.DEV ? 'http://localhost:8285' : '';

// [面向未来] 导出完整的API基础路径 (当前未使用)
export const API_BASE_URL = `${BASE_URL}${API_PREFIX_V1}`;

// [兼容现有] 当前系统使用的API前缀
export const LEGACY_API_PREFIX = '/api';

// ====================================================================
// 单一数据源：所有API端点的集中管理
// 这是所有API路径的唯一真实来源。
// ====================================================================
export const API_ENDPOINTS = {
  auth: {
    login: '/auth/login',
    logout: '/auth/logout',
    refreshToken: '/auth/refresh_token',
    status: '/auth/status',
    changePassword: '/auth/change-password',
    // 2FA endpoints
    setup2FA: '/2fa/setup',
    verify2FA: '/2fa/verify',
    login2FA: '/2fa/login',
    disable2FA: '/2fa/disable',
  },
  users: {
    list: '/users',
    getById: (uid: string) => `/users/${uid}`,
  },
  products: {
    list: '/products',
  },
  examples: {
    list: '/examples',
    create: '/examples',
    getById: (uid: string) => `/examples/${uid}`,
    update: (uid: string) => `/examples/${uid}`,
    delete: (uid: string) => `/examples/${uid}`,
  },
  cli: {
    deviceAuth: {
      sessions: (sessionId: string) => `/cli/device-auth/sessions/${sessionId}`,
      confirm: '/cli/device-auth/confirm',
    },
  },
  containers: {
    list: '/containers',
    checkEnv: '/containers/check-env',
  },
  environments: {
    check: '/environment/check',
    installPodman: '/environment/install-podman',
    installCaddy: '/environment/install-caddy',
  },
  projects: {
    list: '/projects',
    create: '/projects',
    getByUid: (uid: string) => `/projects/${uid}`,
    getByName: (name: string) => `/projects/by-name/${name}`,
    // New application endpoints
    createApp: (uid: string) => `/projects/${uid}/apps`,
    listApps: (uid: string) => `/projects/${uid}/apps`,
    listAppsByName: (name: string) => `/projects/by-name/${name}/apps`,  // 新增：通过名称列出应用
    getAppByName: (projectName: string, appName: string) => `/projects/by-name/${projectName}/apps/by-name/${appName}`, // 新增：通过项目名和应用名获取应用
  },
  applications: {
    create: () => '/apps',
    getById: (uid: string) => `/apps/${uid}`,
    getByName: (name: string) => `/apps/by-name/${name}`,
    update: (uid: string) => `/apps/${uid}`,
    delete: (uid: string) => `/apps/${uid}`,
    status: (uid: string) => `/apps/${uid}/status`,
    logs: (uid: string) => `/apps/${uid}/logs`,
    deployments: (uid: string) => `/apps/${uid}/deployments`,
    runningDeployments: (identifier: string) => `/apps/${identifier}/deployments/running`,
    releases: (uid: string) => `/apps/${uid}/releases`,
    latestRelease: (uid: string) => `/apps/${uid}/releases/latest`,
    configurations: (uid: string) => `/apps/${uid}/configurations`,
    routings: (uid: string) => `/apps/${uid}/routings`,
    // New configuration with environment variables endpoints
    // Application token endpoints
    tokens: (uid: string) => `/apps/${uid}/tokens`,
    tokenCreate: (uid: string) => `/apps/${uid}/tokens`,
    tokenUpdate: (uid: string, tokenId: string) => `/apps/${uid}/tokens/${tokenId}`,
    tokenDelete: (uid: string, tokenId: string) => `/apps/${uid}/tokens/${tokenId}`,
    environmentVariables: (uid: string) => `/apps/${uid}/environment-variables`,
    environmentVariableCreate: (uid: string) => `/apps/${uid}/environment-variables`,
    environmentVariableUpdate: (envVarId: string) => `/environment-variables/${envVarId}`,
    environmentVariableDelete: (envVarId: string) => `/environment-variables/${envVarId}`,
  },
  deployments: {
    getById: (uid: string) => `/deployments/${uid}`,
    logs: (uid: string) => `/deployments/${uid}/logs`,
    restart: (uid: string) => `/deployments/${uid}/restart`,
    status: (uid: string) => `/deployments/${uid}/status`,
  },
  routings: {
    create: (uid: string) => `/apps/${uid}/routings`,
    list: (uid: string) => `/apps/${uid}/routings`,
    update: (routingId: string) => `/routings/${routingId}`,
    delete: (routingId: string) => `/routings/${routingId}`,
  },
  githubTokens: {
    list: '/github-tokens',
    create: '/github-tokens',
    getById: (uid: string) => `/github-tokens/${uid}`,
    update: (uid: string) => `/github-tokens/${uid}`,
    delete: (uid: string) => `/github-tokens/${uid}`,
    test: (uid: string) => `/github-tokens/${uid}/test`,
  },
  providerAuths: {
    list: '/provider-auths',
    create: '/provider-auths',
    getById: (uid: string) => `/provider-auths/${uid}`,
    update: (uid: string) => `/provider-auths/${uid}`,
    delete: (uid: string) => `/provider-auths/${uid}`,
    activate: (uid: string) => `/provider-auths/${uid}/activate`,
    deactivate: (uid: string) => `/provider-auths/${uid}/deactivate`,
    repositories: (uid: string) => `/provider-auths/${uid}/repositories`,
    branches: (uid: string) => `/provider-auths/${uid}/repositories/branches`,
    // GitHub Apps integration endpoints
    githubAppManifest: '/providers/github/app-manifest',
    githubAppCallback: '/providers/github/app-callback',
    githubInstall: (uid: string) => `/provider-auths/${uid}/github-install`,
  },
  sshHosts: {
    list: '/ssh-hosts',
    create: '/ssh-hosts',
    getById: (uid: string) => `/ssh-hosts/${uid}`,
    update: (uid: string) => `/ssh-hosts/${uid}`,
    delete: (uid: string) => `/ssh-hosts/${uid}`,
    test: (uid: string) => `/ssh-hosts/${uid}/test`,
  },
  images: {
    buildFromGitHub: '/images/build-from-github',
  },
  setup: {
    check: '/setup/check',
  },
  system: {
    monitor: '/system/monitor',
  },
  // ... 其他模块
};

// ====================================================================
// URL 构造工具函数
// ====================================================================

/**
 * 构造完整的API URL。
 * 此函数负责拼接传统API前缀和从API_ENDPOINTS获取的路径。
 * @param endpoint - 相对路径，例如 '/examples/1'
 * @returns 返回拼接了前缀的完整路径，例如 '/api/examples/1'
 */
export function buildApiUrl(endpoint: string): string {
  // 确保endpoint以'/'开头
  const normalizedEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
  // 拼接当前系统使用的前缀
  // 生产环境中直接使用相对路径，开发环境依赖Vite代理
  return `${LEGACY_API_PREFIX}${normalizedEndpoint}`;
}

// ====================================================================
// 模块化便捷函数 (推荐使用)
// 它们封装了URL构造逻辑，提供了类型安全的调用方式。
// ====================================================================

// [已优化] 针对examples模块的便捷函数 - 已完成，保持不变
export function getExamplesApiUrl(action: 'list' | 'create' | { type: 'getById' | 'update' | 'delete', uid: string }): string {
  let endpoint: string;

  if (action === 'list') {
    endpoint = API_ENDPOINTS.examples.list;
  } else if (action === 'create') {
    endpoint = API_ENDPOINTS.examples.create;
  } else {
    endpoint = API_ENDPOINTS.examples[action.type](action.uid);
  }

  return buildApiUrl(endpoint);
}

// [无需修改] 针对auth模块的便捷函数 (此为良好实践的例子) - 无 id/uid 逻辑
export function getAuthApiUrl(action: 'login' | 'logout' | 'refreshToken' | 'status' | 'changePassword' | 'setup2FA' | 'verify2FA' | 'login2FA' | 'disable2FA'): string {
  const endpoint = API_ENDPOINTS.auth[action];
  return buildApiUrl(endpoint);
}

// [无需修改] 针对 CLI 模块的便捷函数 (此为良好实践的例子) - 使用 sessionId
export function getCliApiUrl(action: { type: 'sessions', sessionId: string } | 'confirm'): string {
  if (action === 'confirm') {
    return buildApiUrl(API_ENDPOINTS.cli.deviceAuth.confirm);
  } else {
    return buildApiUrl(API_ENDPOINTS.cli.deviceAuth.sessions(action.sessionId));
  }
}

// [无需修改] 针对containers模块的便捷函数 - 无 id/uid 逻辑
export function getContainersApiUrl(action: 'list' | 'checkEnv'): string {
  // 从单一数据源获取路径，而不是硬编码
  const endpoint = API_ENDPOINTS.containers[action];
  return buildApiUrl(endpoint);
}

// [无需修改] 针对environments模块的便捷函数 - 无 id/uid 逻辑
export function getEnvironmentsApiUrl(action: 'check' | 'installPodman' | 'installCaddy'): string {
  const endpoint = API_ENDPOINTS.environments[action];
  return buildApiUrl(endpoint);
}

// [已修复] 针对projects模块的便捷函数
export function getProjectsApiUrl(action: 'list' | 'create' | { type: 'getByUid' | 'createApp' | 'listApps', uid: string } | { type: 'getByName' | 'listAppsByName', name: string } | { type: 'getAppByName', projectName: string, appName: string }): string { 
  let endpoint: string;

  if (action === 'list') {
    endpoint = API_ENDPOINTS.projects.list;
  } else if (action === 'create') {
    endpoint = API_ENDPOINTS.projects.create;
  } else {
    switch (action.type) {
      case 'getByName':
        endpoint = API_ENDPOINTS.projects.getByName(action.name);
        break;
      case 'listAppsByName':
        endpoint = API_ENDPOINTS.projects.listAppsByName(action.name);
        break;
      case 'getAppByName':
        endpoint = API_ENDPOINTS.projects.getAppByName(action.projectName, action.appName);
        break;
      case 'getByUid':
      case 'createApp':
      case 'listApps':
        endpoint = API_ENDPOINTS.projects[action.type](action.uid);
        break;
      default:
        throw new Error('Invalid action type');
    }
  }

  return buildApiUrl(endpoint);
}

// [已修复] 针对applications模块的便捷函数
export function getApplicationsApiUrl(
  action:
    | 'create'
    | { type: 'getById' | 'update' | 'status' | 'logs' | 'deployments' | 'releases' | 'latestRelease' | 'configurations' | 'routings' | 'delete', uid: string }
    | { type: 'runningDeployments', identifier: string }
    | { type: 'getByName', name: string }
    | { type: 'tokens', uid: string, action: 'list' | 'create' }
    | { type: 'tokens', uid: string, action: 'delete', tokenId: string }
    | { type: 'environmentVariables', uid: string }
    | { type: 'environmentVariableCreate', uid: string }
    | { type: 'environmentVariableUpdate' | 'environmentVariableDelete', envVarId: string }
    | { type: 'configurationsWithVariables', uid: string }
): string {
  let endpoint: string;

  if (typeof action === 'string') {
    // Handle the simple string case directly
    endpoint = API_ENDPOINTS.applications[action]();
  } else {
    // Use a switch for the discriminated union based on 'action.type'
    switch (action.type) {
      case 'getById':
      case 'update':
      case 'status':
      case 'logs':
      case 'deployments':
      case 'releases':
      case 'latestRelease':
      case 'configurations':
      case 'routings':
      case 'delete':
        endpoint = API_ENDPOINTS.applications[action.type](action.uid);
        break;

      case 'runningDeployments':
        endpoint = API_ENDPOINTS.applications.runningDeployments(action.identifier);
        break;

      case 'getByName':
        endpoint = API_ENDPOINTS.applications.getByName(action.name);
        break;

      case 'tokens':
        // A nested switch can handle the sub-action cleanly
        switch (action.action) {
          case 'list':
            endpoint = API_ENDPOINTS.applications.tokens(action.uid);
            break;
          case 'create':
            endpoint = API_ENDPOINTS.applications.tokenCreate(action.uid);
            break;
          case 'delete':
            endpoint = API_ENDPOINTS.applications.tokenDelete(action.uid, action.tokenId);
            break;
        }
        break;

      case 'environmentVariables':
      case 'environmentVariableCreate':
        endpoint = API_ENDPOINTS.applications[action.type](action.uid);
        break;

      case 'environmentVariableUpdate':
      case 'environmentVariableDelete':
        endpoint = API_ENDPOINTS.applications[action.type](action.envVarId);
        break;

      case 'configurationsWithVariables':
        // Handle the special mapping
        endpoint = API_ENDPOINTS.applications.configurations(action.uid);
        break;
      
      default:
        // This default case helps with exhaustiveness checking.
        // If a new type is added to the union and not handled here,
        // TypeScript will raise a compile-time error.
        const exhaustiveCheck: never = action;
        throw new Error(`Unhandled action type: ${(exhaustiveCheck as any).type}`);
    }
  }

  return buildApiUrl(endpoint);
}

// [已修复] 针对githubTokens模块的便捷函数
export function getGitHubTokensApiUrl(action: 'list' | 'create' | { type: 'getById' | 'update' | 'delete' | 'test', uid: string }): string {
  let endpoint: string;

  if (action === 'list') {
    endpoint = API_ENDPOINTS.githubTokens.list;
  } else if (action === 'create') {
    endpoint = API_ENDPOINTS.githubTokens.create;
  } else {
    // action.type 是 'getById', 'update', 'delete', 或 'test'
    endpoint = API_ENDPOINTS.githubTokens[action.type](action.uid);
  }

  return buildApiUrl(endpoint);
}

// [已修复] 针对sshHosts模块的便捷函数
export function getSSHHostsApiUrl(action: 'list' | 'create' | { type: 'getById' | 'update' | 'delete' | 'test', uid: string }): string {
  let endpoint: string;

  if (action === 'list') {
    endpoint = API_ENDPOINTS.sshHosts.list;
  } else if (action === 'create') {
    endpoint = API_ENDPOINTS.sshHosts.create;
  } else {
    // action.type 是 'getById', 'update', 'delete', 或 'test'
    endpoint = API_ENDPOINTS.sshHosts[action.type](action.uid);
  }

  return buildApiUrl(endpoint);
}

// [无需修改] 针对images模块的便捷函数
export function getImagesApiUrl(action: 'buildFromGitHub'): string {
  const endpoint = API_ENDPOINTS.images[action];
  return buildApiUrl(endpoint);
}

// [无需修改] 针对setup模块的便捷函数
export function getSetupApiUrl(action: 'check'): string {
  const endpoint = API_ENDPOINTS.setup[action];
  return buildApiUrl(endpoint);
}

// [已修复] 针对routings模块的便捷函数
export function getRoutingsApiUrl(action: { type: 'create' | 'list', uid: string } | { type: 'update' | 'delete', routingId: string }): string {
  let endpoint: string;

  if ('uid' in action) {
    endpoint = API_ENDPOINTS.routings[action.type](action.uid);
  } else {
    endpoint = API_ENDPOINTS.routings[action.type](action.routingId);
  }

  return buildApiUrl(endpoint);
}

// [已修复] 针对environmentVariables模块的便捷函数
export function getEnvironmentVariablesApiUrl(action:
  | { type: 'create' | 'list', uid: string }
  | { type: 'update' | 'delete', envVarId: string }
): string {
  let endpoint: string;

  if ('uid' in action) {
    endpoint = API_ENDPOINTS.applications[action.type === 'create' ? 'environmentVariableCreate' : 'environmentVariables'](action.uid);
  } else {
    endpoint = API_ENDPOINTS.applications[action.type === 'update' ? 'environmentVariableUpdate' : 'environmentVariableDelete'](action.envVarId);
  }

  return buildApiUrl(endpoint);
}

// [已修复] 针对deployments模块的便捷函数
export function getDeploymentsApiUrl(action: { type: 'getById' | 'restart' | 'status' | 'logs', uid: string }): string {
  const endpoint = API_ENDPOINTS.deployments[action.type](action.uid);
  return buildApiUrl(endpoint);
}

// [已修复] 针对providerAuths模块的便捷函数
export function getProviderAuthsApiUrl(
  action:
    | 'list'
    | 'create'
    | 'githubAppManifest'
    | 'githubAppCallback'
    | { type: 'getById' | 'update' | 'delete' | 'activate' | 'deactivate' | 'githubInstall' | 'repositories' | 'branches', uid: string }
): string {
  let endpoint: string;

  if (typeof action === 'string') {
    // Use a switch statement for cleaner handling of string-based actions
    switch (action) {
      case 'list':
        endpoint = API_ENDPOINTS.providerAuths.list;
        break;
      case 'create':
        endpoint = API_ENDPOINTS.providerAuths.create;
        break;
      case 'githubAppManifest':
        endpoint = API_ENDPOINTS.providerAuths.githubAppManifest;
        break;
      case 'githubAppCallback':
        endpoint = API_ENDPOINTS.providerAuths.githubAppCallback;
        break;
    }
  } else {
    endpoint = API_ENDPOINTS.providerAuths[action.type](action.uid);
  }
  return buildApiUrl(endpoint);
}


export const isDev = (): boolean => import.meta.env.DEV;
// eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
export const isProd = (): boolean => !import.meta.env.DEV;