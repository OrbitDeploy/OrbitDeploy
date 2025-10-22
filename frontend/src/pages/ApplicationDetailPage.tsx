import { createSignal, createEffect, Show, Switch, Match, For, Suspense, ErrorBoundary, createMemo } from 'solid-js'
import type { Component } from 'solid-js'
import { useParams, useNavigate, useSearchParams } from '@solidjs/router'
import { useI18n } from '../i18n'
import type { Application, ApplicationLog, Project, DeploymentHistory } from '../types/project'
import { useApiQuery } from '../api/apiHooksW.ts'
import { getProjectByUidEndpoint, getProjectByNameEndpoint, getAppByNameEndpoint, listAppsEndpoint } from '../api/endpoints/projects'
import { useQueryClient } from '@tanstack/solid-query'

// --- Revert lazy imports to regular imports to fix persistent loading on tab switches ---
import ApplicationOverviewTab from '../components/application/ApplicationOverviewTab'
import ApplicationDeploymentsTab from '../components/application/ApplicationDeploymentsTab'
import ApplicationDomainTab from '../components/application/ApplicationDomainTab'
import ApplicationEnvironmentTab from '../components/application/ApplicationEnvironmentTab'
import ApplicationLogsTab from '../components/application/ApplicationLogsTab'
import ApplicationSettingsTab from '../components/application/ApplicationSettingsTab'
import ApplicationTokensTab from '../components/application/ApplicationTokensTab'
import AppSwitcherModal from '../components/application/AppSwitcherModal'

const ApplicationDetailPage: Component = () => {
  const { t } = useI18n()
  const params = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [searchParams, setSearchParams] = useSearchParams()

  // UI state
  const [showAppSwitcherModal, setShowAppSwitcherModal] = createSignal(false)
  
  // State for active tab, with URL query parameter support
  const getInitialTab = () => searchParams.tab || 'Overview'
  const [activeTab, setActiveTab] = createSignal(getInitialTab())

  // Update URL when tab changes
  const handleTabChange = (tab: string) => {
    setActiveTab(tab)
    setSearchParams({ tab }, { replace: true })
  }

  // Get URL parameters - support both name-based and ID-based routing
  const projectIdentifier = () => params.projectuid
  const appIdentifier = () => params.appuid
  
  // Smart detection: check if identifier is a UID (starts with 'prj_' or 'app_')
  const isUid = (str: string | undefined): boolean => {
    if (!str) return false
    return str.startsWith('prj_') || str.startsWith('app_')
  }
  
  const isUidBasedRoute = () => {
    const projId = projectIdentifier()
    const appId = appIdentifier()
    // If either identifier is a UID, treat as UID-based route
    return isUid(projId) || isUid(appId)
  }
  
  const projectUid = () => isUid(projectIdentifier()) ? projectIdentifier() : undefined
  const projectName = () => !isUid(projectIdentifier()) ? projectIdentifier() : undefined
  const appUid = () => isUid(appIdentifier()) ? appIdentifier() : undefined
  const appName = () => !isUid(appIdentifier()) ? appIdentifier() : undefined

  const isNameBasedRoute = () => !!(projectName() && appName())

  // Legacy functions for compatibility
  const projectId = () => {
    return projectIdentifier()
  }

  const appId = () => currentApp()?.uid || null

  // --- 基础数据请求 (页面框架所需，立即加载) ---
  const projectQuery = useApiQuery<Project>(
    () => ['projects', projectIdentifier()],
    () => {
      const pUid = projectUid()
      if (pUid) {
        return getProjectByUidEndpoint(pUid).url
      }
      const pName = projectName()
      if (pName) {
        return getProjectByNameEndpoint(pName).url
      }
      return null
    },
    { enabled: () => !!projectIdentifier() }
  )
  // Computed values
  const currentProject = () => projectQuery.data

  // Direct app query for name-based routes (more efficient)
  const directAppQuery = useApiQuery<Application>(
    () => ['projects', projectIdentifier(), 'apps', appIdentifier()],
    () => {
      const projIdentifier = projectIdentifier()
      const appIdent = appIdentifier()
      
      // Only use direct query for name-based routes
      if (typeof projIdentifier === 'string' && typeof appIdent === 'string') {
        return getAppByNameEndpoint(projIdentifier, appIdent).url
      }
      return null
    },
    { enabled: () => isNameBasedRoute() && !!projectIdentifier() && !!appIdentifier() }
  )

  const projectAppsQuery = useApiQuery<Application[]>(
    () => ['projects', currentProject()?.uid, 'applications'],
    () => {
      const proj = currentProject()
      if (!proj) return null
      return listAppsEndpoint(proj.uid).url
    },
    { enabled: () => !!currentProject() && showAppSwitcherModal() }
  )

  // Computed values
  const projectApps = () => projectAppsQuery.data || []
  const currentApp = () => {
    // For name-based routes, use direct app query result
    if (isNameBasedRoute() && directAppQuery.data) {
      return directAppQuery.data
    }
    
    // For ID-based routes, find from project apps
    const apps = projectApps()
    const identifier = appIdentifier()
    if (!identifier || !apps.length) return undefined // Changed null to undefined
    if (typeof identifier === 'string') {
      return apps.find(app => app.name === identifier || app.uid === identifier)
    } else {
      return apps.find(app => app.uid === identifier)
    }
  }

  // Navigation handlers
  const handleBack = () => {
    const projIdentifier = projectIdentifier()
    if (projIdentifier) {
      if (typeof projIdentifier === 'string') {
        navigate(`/projects/${projIdentifier}`)
      } else {
        navigate(`/projects/${projIdentifier}`)
      }
    } else {
      navigate('/projects')
    }
  }

  // Effect to handle invalid IDs
  createEffect(() => {
    const pId = projectId()
    const aId = appId()
    
    if (projectQuery.isPending || 
        (isNameBasedRoute() ? directAppQuery.isPending : projectAppsQuery.isPending)) return
    
    // if (!pId || !currentApp()) {
    //   navigate('/projects')
    //   return
    // }
    
    if (projectQuery.isError || 
        (isNameBasedRoute() ? directAppQuery.isError : projectAppsQuery.isError)) {
      navigate('/projects')
    }
  })

  // Memoize props to stabilize them and prevent unnecessary re-renders

  const memoizedCurrentApp = createMemo(() => currentApp())

  return (
    <>
      <Switch>
        {/* Loading state */}
        <Match when={projectQuery.isPending || 
                      (isNameBasedRoute() ? directAppQuery.isPending : projectAppsQuery.isPending)}>
          <div class="container mx-auto p-6">
            <div class="flex justify-center py-12">
              <span class="loading loading-spinner loading-lg"></span>
              <p class="mt-4 text-base-content/70 ml-4">{t('common.loading')}</p>
            </div>
          </div>
        </Match>

        {/* Error state */}
        <Match when={projectQuery.isError || 
                      (isNameBasedRoute() ? directAppQuery.isError : projectAppsQuery.isError) || 
                      !currentApp() || !currentProject()}>
          <div class="container mx-auto p-6">
            <div class="alert alert-error mb-4">
              <span>应用或项目未找到</span>
              <button class="btn btn-ghost btn-sm" onClick={handleBack}>返回项目详情</button>
            </div>
          </div>
        </Match>

        {/* Success state */}
        <Match when={true}>
          <div class="container mx-auto p-6">
            {/* Breadcrumbs */}
            <div class="breadcrumbs text-sm mb-6">
              <ul>
                <li><a href="#" onClick={(e) => { e.preventDefault(); navigate('/projects') }}>项目管理</a></li>
                <li><a href="#" onClick={(e) => { e.preventDefault(); handleBack() }}>{currentProject()?.name}</a></li>
                <li class="flex items-center gap-2">
                  {currentApp()?.name}
                  <button
                    class="btn btn-outline btn-xs"
                    onClick={() => setShowAppSwitcherModal(true)}
                  >
                    切换应用
                  </button>
                </li>
              </ul>
            </div>

            {/* Header with App Switcher */}
            {/* <div class="flex justify-between items-center mb-6">
              <div class="flex items-center gap-4">
                <h1 class="text-2xl font-bold">{currentApp()?.name}</h1>
              </div>
              
              <button class="btn btn-outline" onClick={handleBack}>
                返回项目
              </button>
            </div> */}

            {/* Tabs using buttons and state */}
            <div role="tablist" class="tabs tabs-bordered">
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Overview' }}
                onClick={() => handleTabChange('Overview')}
              >
                Overview
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Deployments' }}
                onClick={() => handleTabChange('Deployments')}
              >
                Recent Deployments
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Domain' }}
                onClick={() => handleTabChange('Domain')}
              >
                Domain
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Environment' }}
                onClick={() => handleTabChange('Environment')}
              >
                Environment
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Logs' }}
                onClick={() => handleTabChange('Logs')}
              >
                Logs
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Tokens' }}
                onClick={() => handleTabChange('Tokens')}
              >
                Tokens
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Settings' }}
                onClick={() => handleTabChange('Settings')}
              >
                Settings
              </button>
            </div>

            {/* Tab Content Area */}
            <div class="p-4 border-base-300 border border-t-0 rounded-b-lg">
              <ErrorBoundary fallback={(err) => <div class="alert alert-error">Error loading tab: {err.message}</div>}>
                <Suspense fallback={<span class="loading loading-dots loading-md"></span>}>
                  <Switch>
                    {/* --- All tab components now imported regularly, no lazy loading --- */}
                    <Match when={activeTab() === 'Overview'}>
                      <ApplicationOverviewTab applicationUid={appId()!} currentApp={currentApp()} appIdentifier={appIdentifier()} />
                    </Match>
                    <Match when={activeTab() === 'Deployments'}>
                      <ApplicationDeploymentsTab applicationUid={appId()!} />
                    </Match>
                    <Match when={activeTab() === 'Domain'}>
                      <ApplicationDomainTab applicationUid={appId()!} />
                    </Match>
                    <Match when={activeTab() === 'Environment'}>
                      <ApplicationEnvironmentTab applicationUid={appId()!} />
                    </Match>
                    <Match when={activeTab() === 'Logs'}>
                      <ApplicationLogsTab applicationUid={appId()!} />
                    </Match>
                    <Match when={activeTab() === 'Tokens'}>
                      <ApplicationTokensTab application={currentApp()!} />
                    </Match>
                    <Match when={activeTab() === 'Settings'}>
                      <ApplicationSettingsTab currentApp={currentApp()} />
                    </Match>
                  </Switch>
                </Suspense>
              </ErrorBoundary>
            </div>
          </div>
        </Match>
      </Switch>

      {/* App Switcher Modal */}
      <AppSwitcherModal
        isOpen={showAppSwitcherModal()}
        currentAppUid={appId() || ''}
        projectUid={projectId() || ''}
        projectName={currentProject()?.name}
        applications={projectApps()}
        onClose={() => setShowAppSwitcherModal(false)}
      />
    </>
  )
}

export default ApplicationDetailPage