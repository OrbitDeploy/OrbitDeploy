import { createSignal, Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useQueryClient } from '@tanstack/solid-query'
import { toast } from 'solid-toast'
import { useApiQuery, useApiMutation } from '../api/apiHooksW.ts'
import { listSshHostsEndpoint, createSshHostEndpoint, updateSshHostEndpoint, deleteSshHostEndpoint } from '../api/endpoints/sshHosts'
import { useI18n } from '../i18n'
import RemoteContainerManagement from '../components/RemoteContainerManagement'
import SSHTerminalModal from '../components/SSHTerminalModal'
import CreateSSHHostModal from '../components/ssh/CreateSSHHostModal'
import EditSSHHostModal from '../components/ssh/EditSSHHostModal'
import DeleteSSHHostModal from '../components/ssh/DeleteSSHHostModal'
import SSHHostTable from '../components/ssh/SSHHostTable'
import type { SSHHost, SSHHostRequest } from '../types/remote'

interface SSHHostApiListResponse {
  data: SSHHost[]
}

const SSHManagementPage: Component = () => {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  
  // Modal states
  const [showCreateModal, setShowCreateModal] = createSignal(false)
  const [showEditModal, setShowEditModal] = createSignal(false)
  const [showDeleteModal, setShowDeleteModal] = createSignal(false)
  const [showSSHModal, setShowSSHModal] = createSignal(false)
  const [selectedHost, setSelectedHost] = createSignal<SSHHost | null>(null)

  // Tab state for showing SSH hosts or remote containers
  const [activeTab, setActiveTab] = createSignal<'hosts' | 'containers'>('hosts')

  // API query for SSH hosts
  const hostsQuery = useApiQuery<SSHHostApiListResponse>(
    ['ssh-hosts'],
    () => listSshHostsEndpoint().url
  )

  const refreshHosts = async () => {
    await queryClient.invalidateQueries({ queryKey: ['ssh-hosts'] })
  }

  // API mutations
  const createHostMutation = useApiMutation<unknown, SSHHostRequest>(
    createSshHostEndpoint(),
    {
      onSuccess: () => {
        setShowCreateModal(false)
        toast.success('SSH host created successfully')
        void refreshHosts()
      },
      onError: (error: Error) => {
        toast.error(error.message || 'Failed to create SSH host')
      }
    }
  )

  const updateHostMutation = useApiMutation<unknown, { uid: string; data: SSHHostRequest }>(
    (variables) => updateSshHostEndpoint(variables.uid),
    {
      body: (variables) => variables.data,
      onSuccess: () => {
        setShowEditModal(false)
        setSelectedHost(null)
        toast.success('SSH host updated successfully')
        void refreshHosts()
      },
      onError: (error: Error) => {
        toast.error(error.message || 'Failed to update SSH host')
      }
    }
  )

  const deleteHostMutation = useApiMutation<unknown, { uid: string }>(
    (variables) => deleteSshHostEndpoint(variables.uid),
    {
      onSuccess: () => {
        setShowDeleteModal(false)
        setSelectedHost(null)
        toast.success('SSH host deleted successfully')
        void refreshHosts()
      },
      onError: (error: Error) => {
        toast.error(error.message || 'Failed to delete SSH host')
      }
    }
  )

  // Helper functions
  const hosts = () => hostsQuery.data || []
  const isLoading = () => hostsQuery.isPending
  const isMutating = () => createHostMutation.isPending || updateHostMutation.isPending || deleteHostMutation.isPending

  const createHost = (data: SSHHostRequest) => {
    if (!data.name || !data.addr || !data.user) {
      toast.error('Name, address, and user are required')
      return
    }

    if (!data.password && !data.private_key) {
      toast.error('Either password or private key must be provided')
      return
    }

    createHostMutation.mutate(data)
  }

  const updateHost = (data: SSHHostRequest) => {
    const host = selectedHost()
    if (!host) return

    if (!data.name || !data.addr || !data.user) {
      toast.error('Name, address, and user are required')
      return
    }

    if (!data.password && !data.private_key) {
      toast.error('Either password or private key must be provided')
      return
    }

    updateHostMutation.mutate({ uid: host.uid, data })
  }

  const deleteHost = () => {
    const host = selectedHost()
    if (!host) return

    deleteHostMutation.mutate({ uid: host.uid })
  }

  const connectSSH = (host: SSHHost) => {
    setSelectedHost(host)
    setShowSSHModal(true)
  }

  const openEditModal = (host: SSHHost) => {
    setSelectedHost(host)
    setShowEditModal(true)
  }

  const openDeleteModal = (host: SSHHost) => {
    setSelectedHost(host)
    setShowDeleteModal(true)
  }

  return (
    <div class="container mx-auto p-6">
      {/* Header */}
      <div class="mb-6 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-bold text-base-content">{t('ssh.title')}</h1>
          <p class="text-base-content/70 mt-2">{t('ssh.description')}</p>
        </div>
        
        <Show when={activeTab() === 'hosts'}>
          <button 
            class="btn btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            {t('ssh.add_host')}
          </button>
        </Show>
      </div>

      {/* Tab Navigation */}
      <div class="tabs tabs-bordered mb-6">
        <button 
          class={`tab tab-lg ${activeTab() === 'hosts' ? 'tab-active' : ''}`}
          onClick={() => setActiveTab('hosts')}
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3V6a3 3 0 013-3h13.5a3 3 0 013 3v5.25a3 3 0 01-3 3m-16.5 0a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25h15a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25h-16.5z" />
          </svg>
          SSH Hosts ({hosts().length})
        </button>
        <button 
          class={`tab tab-lg ${activeTab() === 'containers' ? 'tab-active' : ''}`}
          onClick={() => setActiveTab('containers')}
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 mr-2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 18.847 16.556 20.625 12 20.625s-8.25-1.778-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
          </svg>
          Remote Containers
        </button>
      </div>

      {/* Tab Content */}
      <Show when={activeTab() === 'hosts'}>
        <SSHHostTable
          hosts={hosts()}
          isLoading={isLoading()}
          onConnect={connectSSH}
          onEdit={openEditModal}
          onDelete={openDeleteModal}
        />
      </Show>

      <Show when={activeTab() === 'containers'}>
        <RemoteContainerManagement 
          hosts={hosts()} 
          onRefreshHosts={() => void refreshHosts()}
        />
      </Show>

      {/* Modals */}
      <CreateSSHHostModal
        isOpen={showCreateModal()}
        onClose={() => setShowCreateModal(false)}
        onSubmit={createHost}
        isLoading={isMutating()}
      />

      <EditSSHHostModal
        isOpen={showEditModal()}
        host={selectedHost()}
        onClose={() => {
          setShowEditModal(false)
          setSelectedHost(null)
        }}
        onSubmit={updateHost}
        isLoading={isMutating()}
      />

      <DeleteSSHHostModal
        isOpen={showDeleteModal()}
        host={selectedHost()}
        onClose={() => {
          setShowDeleteModal(false)
          setSelectedHost(null)
        }}
        onConfirm={deleteHost}
        isLoading={isMutating()}
      />

      <SSHTerminalModal 
        host={selectedHost()}
        isOpen={showSSHModal()}
        onClose={() => {
          setShowSSHModal(false)
          setSelectedHost(null)
        }}
        token={(window as any).getAccessToken?.()}
      />
    </div>
  )
}

export default SSHManagementPage