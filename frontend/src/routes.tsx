import { lazy } from 'solid-js'
import type { RouteDefinition } from '@solidjs/router'
import SetupPageWithRouter from './components/SetupPageWithRouter'
import LoginPage from './pages/LoginPage'
import ProtectedRoute from './components/ProtectedRoute'

// Lazy load pages for better performance
const DashboardPage = lazy(() => import('./pages/DashboardPage'))
const SSHManagementPage = lazy(() => import('./pages/SSHManagementPage'))
const ProjectListPage = lazy(() => import('./pages/ProjectListPage'))
const ProjectDetailPage = lazy(() => import('./pages/ProjectDetailPage'))
const ApplicationDetailPage = lazy(() => import('./pages/ApplicationDetailPage'))

const ChangePasswordPage = lazy(() => import('./pages/ChangePasswordPage'))
const SystemSettingsPage = lazy(() => import('./pages/SystemSettingsPage'))
const SystemMonitorPage = lazy(() => import('./pages/SystemMonitorPage'))
const CLIAuthorizePage = lazy(() => import('./pages/CLIAuthorizePage'))
const CLIProjectConfigPage = lazy(() => import('./pages/CLIProjectConfigPage'))
const CLIDeviceAuthPage = lazy(() => import('./pages/CLIDeviceAuthPage'))
const GitHubTokenManagementPage = lazy(() => import('./pages/GitHubTokenManagementPage'))
const ProviderAuthManagementPage = lazy(() => import('./pages/ProviderAuthManagementPage'))
const DatabaseManagementPage = lazy(() => import('./pages/DatabaseManagementPage'))

export const routes: RouteDefinition[] = [
  {
    path: '/login',
    component: LoginPage,
  },
  {
    path: '/setup',
    component: SetupPageWithRouter,
  },
  {
    path: '/cli-authorize',
    component: () => (
      <ProtectedRoute>
        <CLIAuthorizePage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/cli-device-auth',
    component: () => (
      <ProtectedRoute>
        <CLIDeviceAuthPage />
      </ProtectedRoute>
    ),
  },

  {
    path: '/cli-configure',
    component: CLIProjectConfigPage, // No authentication required for this temp page
  },
  {
    path: '/',
    component: () => (
      <ProtectedRoute>
        <DashboardPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/dashboard',
    component: () => (
      <ProtectedRoute>
        <DashboardPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/containers',
    component: () => (
      <ProtectedRoute>
        <ContainerManagementPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/ssh-management',
    component: () => (
      <ProtectedRoute>
        <SSHManagementPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/docker-images',
    component: () => (
      <ProtectedRoute>
        <DockerImageManagementPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/projects',
    component: () => (
      <ProtectedRoute>
        <ProjectListPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/projects/:projectuid/apps/:appuid',  
    component: () => (
      <ProtectedRoute>
        <ApplicationDetailPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/projects/:projectuid',  // Unified route - supports both UID and name
    component: () => (
      <ProtectedRoute>
        <ProjectDetailPage />
      </ProtectedRoute>
    ),
  },

  {
    path: '/examples-query',
    component: () => (
      <ProtectedRoute>
        <ExamplesPageQuery />
      </ProtectedRoute>
    ),
  },
  {
    path: '/github-tokens',
    component: () => (
      <ProtectedRoute>
        <GitHubTokenManagementPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/provider-auths',
    component: () => (
      <ProtectedRoute>
        <ProviderAuthManagementPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/databases',
    component: () => (
      <ProtectedRoute>
        <DatabaseManagementPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/change-password',
    component: () => (
      <ProtectedRoute>
        <ChangePasswordPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/settings/system',
    component: () => (
      <ProtectedRoute>
        <SystemSettingsPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/system-monitor',
    component: () => (
      <ProtectedRoute>
        <SystemMonitorPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/*all',
    component: () => (
      <ProtectedRoute>
        <DashboardPage />
      </ProtectedRoute>
    ),
  },
]