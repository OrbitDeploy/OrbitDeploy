import { Component, Show, JSX } from 'solid-js'
import { Navigate, useLocation } from '@solidjs/router'
import { useAuth } from '../contexts/AuthContext'
import { useApiQuery } from '../lib/apiHooks'
import { getSetupApiUrl } from '../api/config'
import AdminLayout from './AdminLayout'

interface ProtectedRouteProps {
  children: JSX.Element
}

const ProtectedRoute: Component<ProtectedRouteProps> = (props) => {
  const auth = useAuth()
  const location = useLocation()
  
  // Check if setup is required using the new API management system
  const setupQuery = useApiQuery<{ setup_required: boolean }>(
    ['setup', 'check'],
    getSetupApiUrl('check')
  )

  return (
    <Show
      when={!auth.isLoading()}
      fallback={
        <div class="min-h-screen bg-base-200 flex items-center justify-center">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
      }
    >
      <Show
        when={!setupQuery.isLoading}
        fallback={
          <div class="min-h-screen bg-base-200 flex items-center justify-center">
            <span class="loading loading-spinner loading-lg"></span>
          </div>
        }
      >
        <Show
          when={!setupQuery.data?.setup_required}
          fallback={<Navigate href="/setup" />}
        >
          <Show
            when={auth.isAuthenticated()}
            fallback={<Navigate href="/login" />}
          >
            <AdminLayout>
              {props.children}
            </AdminLayout>
          </Show>
        </Show>
      </Show>
    </Show>
  )
}

export default ProtectedRoute