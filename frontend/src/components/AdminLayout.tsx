import { createSignal, JSX } from 'solid-js'
import { useNavigate, useLocation } from '@solidjs/router'
import { useAuth } from '../contexts/AuthContext'
import { useI18n } from '../i18n'
import LanguageSwitcher from './LanguageSwitcher'

interface AdminLayoutProps {
  children: JSX.Element
}

const AdminLayout = (props: AdminLayoutProps) => {
  const auth = useAuth()
  const { t } = useI18n()
  const navigate = useNavigate()
  const location = useLocation()
  const [sidebarOpen, setSidebarOpen] = createSignal(false)

  const handleLogout = () => {
    void auth.logout()
  }

  // Get current page from the current route
  const getCurrentPage = () => {
    const path = location.pathname
    if (path === '/' || path === '/dashboard') return 'dashboard'
    return path.substring(1) // Remove leading slash
  }

  const menuItems = [
    { id: 'dashboard', name: t('nav.dashboard'), href: '/dashboard', icon: 'M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z' },
    // { id: 'containers', name: t('nav.deployment'), href: '/containers', icon: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16' },
    { id: 'ssh-management', name: t('nav.ssh_management'), href: '/ssh-management', icon: 'M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z' },
    // { id: 'docker-images', name: t('nav.docker_images'), href: '/docker-images', icon: 'M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7h16zM9 6a1 1 0 100 2 1 1 0 000-2zM6 8a1 1 0 100 2 1 1 0 000-2zM6 10a1 1 0 100 2 1 1 0 000-2zM18 6a1 1 0 100 2 1 1 0 000-2zM15 6a1 1 0 100 2 1 1 0 000-2zM12 6a1 1 0 100 2 1 1 0 000-2z' },
    { id: 'projects', name: t('nav.projects') || '项目管理', href: '/projects', icon: 'M3 7h18M3 12h18M3 17h18' },
    { id: 'databases', name: t('nav.database_management') || '数据库管理', href: '/databases', icon: 'M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.653 16.556 18.375 12 18.375s-8.25-1.722-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125' },
    // { id: 'github-tokens', name: 'GitHub令牌', href: '/github-tokens', icon: 'M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z' },
    { id: 'provider-auths', name: t('nav.provider_auths'), href: '/provider-auths', icon: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z' },
    { id: 'system_monitor', name: t('nav.system_monitor'), href: '/system-monitor', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
  ]

  return (
    <div class="drawer lg:drawer-open">
      <input 
        id="drawer-toggle" 
        type="checkbox" 
        class="drawer-toggle" 
        checked={sidebarOpen()}
        onChange={(e) => setSidebarOpen(e.target.checked)}
      />
      
      <div class="drawer-content flex flex-col">
        {/* Top Navigation */}
        <div class="navbar bg-base-100 border-b border-base-300">
          <div class="flex-none lg:hidden">
            <label for="drawer-toggle" class="btn btn-square btn-ghost">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-6 h-6 stroke-current">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
              </svg>
            </label>
          </div>
          
          <div class="flex-1">
            <a class="btn btn-ghost text-xl">{t('nav.admin_title')}</a>
          </div>
          
          <div class="flex-none">
            <LanguageSwitcher />
            <div class="dropdown dropdown-end ml-2">
              <div tabindex="0" role="button" class="btn btn-ghost btn-circle avatar">
                <div class="w-10 rounded-full bg-base-300 flex items-center justify-center">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-6 h-6 stroke-current">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                  </svg>
                </div>
              </div>
              <ul tabindex="0" class="mt-3 z-[1] p-2 shadow menu menu-sm dropdown-content bg-base-100 rounded-box w-52">
                <li class="menu-title">
                  <span>{auth.user()?.username}</span>
                </li>
                <li><a onClick={() => navigate('/change-password')}>{t('nav.change_password')}</a></li>
                <li><a onClick={() => navigate('/settings/system')}>{t('nav.system_settings')}</a></li>
                <li><a onClick={handleLogout}>{t('nav.logout')}</a></li>
              </ul>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <main class="flex-1 overflow-y-auto bg-base-200">
          {props.children}
        </main>
      </div>
      
      {/* Sidebar */}
      <div class="drawer-side">
        <label for="drawer-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
        <aside class="min-h-full w-64 bg-base-100 border-r border-base-300">
          <div class="p-4">
            <div class="text-xl font-bold text-center">{t('nav.system_title')}</div>
          </div>
          
          <ul class="menu p-4 w-full">
            {menuItems.map(item => (
              <li>
                <a 
                  class={`${getCurrentPage() === item.uid ? 'active' : ''}`}
                  onClick={() => {
                    navigate(item.href)
                    setSidebarOpen(false)
                  }}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-5 h-5 stroke-current">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon}></path>
                  </svg>
                  {item.name}
                </a>
              </li>
            ))}
          </ul>
        </aside>
      </div>
    </div>
  )
}

export default AdminLayout