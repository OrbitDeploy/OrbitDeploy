import { createSignal, createEffect, onCleanup, Switch, Match, Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useParams, useNavigate } from '@solidjs/router'
import { useI18n } from '../i18n'
import type { Project } from '../types/project'
import ProjectDetailsView from '../components/project/ProjectDetailsView'
import ApplicationList from '../components/ApplicationList'
import CreateApplicationModal from '../components/application/CreateApplicationModal'
import { useQueryClient } from '@tanstack/solid-query'
import { useApiQuery } from '../api/apiHooksW.ts'
import { getProjectByUidEndpoint, getProjectByNameEndpoint } from '../api/endpoints/projects'

const ProjectDetailPage: Component = () => {
  const { t } = useI18n()
  const params = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()


  // UI state
  const [showCreateApplicationModal, setShowCreateApplicationModal] = createSignal(false)
  const [error, setError] = createSignal('')
  const [success, setSuccess] = createSignal('')  
  
  // Get the project identifier from URL - could be UID or name
  const projectIdentifier = () => params.projectuid
  
  // Smart detection: check if identifier is a UID (starts with 'prj_')
  const isUid = (str: string | undefined): boolean => {
    if (!str) return false
    return str.startsWith('prj_')
  }
  
  const isUidBasedRoute = () => {
    const identifier = projectIdentifier()
    return isUid(identifier)
  }
  
  const projectUid = () => isUidBasedRoute() ? projectIdentifier() : undefined
  const projectName = () => !isUidBasedRoute() ? projectIdentifier() : undefined

  // 使用 TanStack Query 加载项目详情
  const projectQuery = useApiQuery<Project>(
    () => ['projects', projectIdentifier()],
    () => {
      const uid = projectUid()
      if (uid) {
          return getProjectByUidEndpoint(uid).url
      } else {
        return getProjectByNameEndpoint(projectName()!).url
      }
    },
    {
      enabled: () => !!projectIdentifier(),
    }
  )


  // 当前项目访问器
  const currentProject = () => projectQuery.data

  // 导航回项目列表
  function handleBack() {
    navigate('/projects')
  }


  

  function handleCreateApplication() {
    setShowCreateApplicationModal(true)
  }






 
  // // 添加页面事件监听器用于调试刷新和导航问题
  // createEffect(() => {
  //   const handleBeforeUnload = (event: BeforeUnloadEvent) => {
  //     console.log('项目详情页: 页面即将卸载（刷新/导航）', event)
  //   }

  //   const handleVisibilityChange = () => {
  //     console.log('项目详情页: 可见性变化，隐藏状态:', document.hidden)
  //   }

  //   const handlePopState = (event: PopStateEvent) => {
  //     console.log('项目详情页: PopState事件（前进/后退导航）', event)
  //   }

  //   window.addEventListener('beforeunload', handleBeforeUnload)
  //   document.addEventListener('visibilitychange', handleVisibilityChange)
  //   window.addEventListener('popstate', handlePopState)

  //   onCleanup(() => {
  //     console.log('项目详情页: 移除页面事件监听器')
  //     window.removeEventListener('beforeunload', handleBeforeUnload)
  //     document.removeEventListener('visibilitychange', handleVisibilityChange)
  //     window.removeEventListener('popstate', handlePopState)
  //   })
  // })

  // 处理无效的项目ID或名称
  createEffect(() => {
    const identifier = projectIdentifier()
    
    // 当查询还在进行中时，不执行重定向逻辑
    if (projectQuery.isPending) return

    if (!identifier) {
      navigate('/projects')
      return
    }
    
    if (projectQuery.isError) {
      console.error('项目详情页: 项目查询失败，重定向到项目列表。错误:', projectQuery.error)
      setError('项目未找到')
      navigate('/projects')
    }
  })

  return (
    <Switch>
      {/* 匹配1：加载状态 */}
      <Match when={projectQuery.isPending}>
        <div class="container mx-auto p-6">
          <div class="flex justify-center py-12">
            <span class="loading loading-spinner loading-lg"></span>
            <p class="mt-4 text-base-content/70 ml-4">{t('common.loading')}</p>
          </div>
        </div>
      </Match>

      {/* 匹配2：错误状态 */}
      <Match when={projectQuery.isError}>
        <div class="container mx-auto p-6">
          <div class="alert alert-error mb-4">
            <span>项目加载失败或未找到</span>
            <div class="flex gap-2">
              <button 
                class="btn btn-ghost btn-sm" 
                onClick={() => {
                  void projectQuery.refetch()
                }}
              >
                重试
              </button>
              <button class="btn btn-ghost btn-sm" onClick={handleBack}>返回项目列表</button>
            </div>
          </div>
        </div>
      </Match>

      {/* 匹配3：无数据状态 */}
      <Match when={!currentProject()}>
        <div class="container mx-auto p-6">
          <div class="alert alert-warning mb-4">
            <span>项目数据不存在</span>
            <div class="flex gap-2">
              <button class="btn btn-ghost btn-sm" onClick={handleBack}>返回项目列表</button>
            </div>
          </div>
        </div>
      </Match>

      {/* 默认情况 (Fallback)：成功状态 */}
      <Match when={true}>
        <div class="container mx-auto p-6">
          {/* 面包屑导航 */}
          <div class="breadcrumbs text-sm mb-6">
            <ul>
              <li><a href="#" onClick={(e) => { e.preventDefault(); handleBack() }}>项目管理</a></li>
              <li>{currentProject()!.name}</li>
            </ul>
          </div>

          {error() && (
            <div class="alert alert-error mb-4">
              <span>{error()}</span>
              <button class="btn btn-ghost btn-sm" onClick={() => setError('')}>×</button>
            </div>
          )}
          {success() && (
            <div class="alert alert-success mb-4">
              <span>{success()}</span>
              <button class="btn btn-ghost btn-sm" onClick={() => setSuccess('')}>×</button>
            </div>
          )}

          {/* Project detail view */}
          <ProjectDetailsView
            project={currentProject()!}
            onBack={handleBack}
          />

          {/* Application List in Overview */}
          <Show when={currentProject()}>
            <ApplicationList
              projectUid={currentProject()!.uid}
              projectName={currentProject()!.name}
              onCreateApplication={handleCreateApplication}
            />
          </Show>


      
        


          {/* Create Application Modal */}
          <CreateApplicationModal
            isOpen={showCreateApplicationModal()}
            projectUid={currentProject()?.uid || ''}
            projectName={currentProject()?.name || ''}
            onClose={() => setShowCreateApplicationModal(false)}
            onSuccess={setSuccess}
            onError={setError}
            onRefresh={() => {
              void queryClient.invalidateQueries({
                predicate: (query) => query.queryKey[0] === 'projects'
              })
            }}
          />
        </div>
      </Match>
    </Switch>
    
  )
}

export default ProjectDetailPage