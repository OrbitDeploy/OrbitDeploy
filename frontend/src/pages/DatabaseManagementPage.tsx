import { createSignal, Show, For } from 'solid-js'
import type { Component } from 'solid-js'
import { useQueryClient } from '@tanstack/solid-query'
import { toast } from 'solid-toast'
import { useApiQuery, useApiMutation } from '../api/apiHooksW.ts'
import {
  listDatabasesEndpoint,
  createDatabaseEndpoint,
  deleteDatabaseEndpoint,
  deployDatabaseEndpoint,
  startDatabaseEndpoint,
  stopDatabaseEndpoint,
  restartDatabaseEndpoint,
  getDatabaseConnectionInfoEndpoint,
} from '../api/endpoints/databases'
import { useI18n } from '../i18n'
import type {
  Database,
  CreateDatabaseRequest,
  DatabaseConnectionInfo,
} from '../types/database'

interface DatabaseListResponse {
  data: Database[]
}

const DatabaseManagementPage: Component = () => {
  const { t } = useI18n()
  const queryClient = useQueryClient()

  // Modal states
  const [showCreateModal, setShowCreateModal] = createSignal(false)
  const [showDeleteModal, setShowDeleteModal] = createSignal(false)
  const [showConnectionModal, setShowConnectionModal] = createSignal(false)
  const [selectedDatabase, setSelectedDatabase] = createSignal<Database | null>(null)
  const [connectionInfo, setConnectionInfo] = createSignal<DatabaseConnectionInfo | null>(null)
  const [showPassword, setShowPassword] = createSignal(false)

  // Form state
  const generatePort = () => Math.floor(Math.random() * (65535 - 10000 + 1)) + 10000

  const [formData, setFormData] = createSignal<CreateDatabaseRequest>({
    name: '',
    type: 'postgresql',
    version: '16-alpine',
    custom_image: '',
    port: generatePort(),
    internal_port: 5432,
    username: 'postgres',
    password: '',
    database_name: 'postgres',
    data_path: '/var/lib/orbit-deploy/db-data',
    config_path: '',
    is_remote: false,
  })

  // Toggle for custom image input
  const [useCustomImage, setUseCustomImage] = createSignal(false)

  // API query for databases
  const databasesQuery = useApiQuery<DatabaseListResponse>(
    ['databases'],
    () => listDatabasesEndpoint().url
  )

  const refreshDatabases = async () => {
    await queryClient.invalidateQueries({ queryKey: ['databases'] })
  }

  // API mutations
  const createDatabaseMutation = useApiMutation<unknown, CreateDatabaseRequest>(
    createDatabaseEndpoint(),
    {
      onSuccess: () => {
        setShowCreateModal(false)
        toast.success(t('database.create_success'))
        void refreshDatabases()
        resetForm()
      },
      onError: (error: Error) => {
        toast.error(error.message || t('database.create_failed'))
      },
    }
  )

  const deleteDatabaseMutation = useApiMutation<unknown, { uid: string }>(
    (variables) => deleteDatabaseEndpoint(variables.uid),
    {
      onSuccess: () => {
        setShowDeleteModal(false)
        setSelectedDatabase(null)
        toast.success(t('database.delete_success'))
        void refreshDatabases()
      },
      onError: (error: Error) => {
        toast.error(error.message || t('database.delete_failed'))
      },
    }
  )

  const deployDatabaseMutation = useApiMutation<unknown, { uid: string }>(
    (variables) => deployDatabaseEndpoint(variables.uid),
    {
      onSuccess: () => {
        toast.success(t('database.deploy_success'))
        void refreshDatabases()
      },
      onError: (error: Error) => {
        toast.error(error.message || t('database.deploy_failed'))
      },
    }
  )

  const startDatabaseMutation = useApiMutation<unknown, { uid: string }>(
    (variables) => startDatabaseEndpoint(variables.uid),
    {
      onSuccess: () => {
        toast.success(t('database.start_success'))
        void refreshDatabases()
      },
      onError: (error: Error) => {
        toast.error(error.message || t('database.start_failed'))
      },
    }
  )

    const stopDatabaseMutation = useApiMutation<unknown, { uid: string }>(

      (variables) => stopDatabaseEndpoint(variables.uid),

      {

        onSuccess: () => {

          console.log("stop success");

          toast.success(t('database.stop_success'))

          void refreshDatabases()

        },

        onError: (error: Error) => {

          console.error("stop error", error);

          toast.error(error.message || t('database.stop_failed'))

        },

      }

    )

  

    const restartDatabaseMutation = useApiMutation<unknown, { uid: string }>(

      (variables) => restartDatabaseEndpoint(variables.uid),

      {

        onSuccess: () => {

          console.log("restart success");

          toast.success(t('database.restart_success'))

          void refreshDatabases()

        },

        onError: (error: Error) => {

          console.error("restart error", error);

          toast.error(error.message || t('database.restart_failed'))

        },

      }

    )

  // Helper functions
  const databases = () => databasesQuery.data || []
  const isLoading = () => databasesQuery.isPending

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'postgresql',
      version: '16-alpine',
      custom_image: '',
      port: generatePort(),
      internal_port: 5432,
      username: 'postgres',
      password: '',
      database_name: 'postgres',
      data_path: '/var/lib/orbit-deploy/db-data',
      config_path: '',
      is_remote: false,
    })
    setUseCustomImage(false)
  }

  const generatePassword = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*'
    let password = ''
    for (let i = 0; i < 16; i++) {
      password += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    setFormData({ ...formData(), password })
  }

  const handleCreateDatabase = () => {
    const data = formData()
    if (!data.name || !data.password) {
      toast.error('Name and password are required')
      return
    }
    createDatabaseMutation.mutate(data)
  }

  const handleDeleteDatabase = () => {
    const db = selectedDatabase()
    if (!db) return
    deleteDatabaseMutation.mutate({ uid: db.uid })
  }

  const openDeleteModal = (db: Database) => {
    setSelectedDatabase(db)
    setShowDeleteModal(true)
  }

  const handleDeploy = (db: Database) => {
    deployDatabaseMutation.mutate({ uid: db.uid })
  }

  const handleStart = (db: Database) => {
    startDatabaseMutation.mutate({ uid: db.uid })
  }

  const handleStop = (db: Database) => {
    stopDatabaseMutation.mutate({ uid: db.uid })
  }

  const handleRestart = (db: Database) => {
    restartDatabaseMutation.mutate({ uid: db.uid })
  }

  const showConnectionInfo = async (db: Database) => {
    try {
      const endpoint = getDatabaseConnectionInfoEndpoint(db.uid, showPassword())
      const response = await fetch(endpoint.url, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('access_token')}`,
        },
      })
      const result = await response.json()
      if (result.success) {
        setConnectionInfo(result.data)
        setSelectedDatabase(db)
        setShowConnectionModal(true)
      }
    } catch (error) {
      toast.error('Failed to fetch connection info')
    }
  }

  const copyConnectionString = () => {
    const info = connectionInfo()
    if (info) {
      navigator.clipboard.writeText(info.connection_string)
      toast.success(t('database.copied'))
    }
  }

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'running':
        return 'badge-success'
      case 'stopped':
        return 'badge-warning'
      case 'failed':
        return 'badge-error'
      case 'pending':
        return 'badge-info'
      default:
        return 'badge-ghost'
    }
  }

  return (
    <div class="container mx-auto p-6">
      {/* Header */}
      <div class="mb-6 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-bold text-base-content">{t('database.title')}</h1>
          <p class="text-base-content/70 mt-2">{t('database.list_title')}</p>
        </div>

        <button class="btn btn-primary" onClick={() => setShowCreateModal(true)}>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke-width="1.5"
            stroke="currentColor"
            class="w-5 h-5"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          {t('database.create_database')}
        </button>
      </div>

      {/* Database List */}
      <Show
        when={!isLoading()}
        fallback={
          <div class="flex justify-center items-center h-64">
            <span class="loading loading-spinner loading-lg"></span>
          </div>
        }
      >
        <Show
          when={databases().length > 0}
          fallback={
            <div class="card bg-base-100 shadow-xl">
              <div class="card-body items-center text-center">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke-width="1.5"
                  stroke="currentColor"
                  class="w-16 h-16 text-base-content/30 mb-4"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.653 16.556 18.375 12 18.375s-8.25-1.722-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125"
                  />
                </svg>
                <h2 class="card-title">{t('database.empty_state')}</h2>
                <p>{t('database.empty_state_desc')}</p>
              </div>
            </div>
          }
        >
          <div class="overflow-x-auto">
            <table class="table table-zebra w-full">
              <thead>
                <tr>
                  <th>{t('database.database_name')}</th>
                  <th>{t('database.database_type')}</th>
                  <th>{t('database.version')}</th>
                  <th>{t('database.status')}</th>
                  <th>{t('database.port')}</th>
                  <th>{t('database.username')}</th>
                  <th>{t('database.actions')}</th>
                </tr>
              </thead>
              <tbody>
                <For each={databases()}>
                  {(db) => (
                    <tr>
                      <td class="font-medium">{db.name}</td>
                      <td>{t(`database.type_${db.type}`)}</td>
                      <td>{db.version}</td>
                      <td>
                        <span class={`badge ${getStatusBadgeClass(db.status)}`}>
                          {t(`database.status_${db.status}`)}
                        </span>
                      </td>
                      <td>{db.port}</td>
                      <td>{db.username}</td>
                      <td>
                        <div class="flex gap-2">
                          <Show when={db.status === 'pending'}>
                            <button
                              class="btn btn-sm btn-primary"
                              onClick={() => handleDeploy(db)}
                              disabled={deployDatabaseMutation.isPending}
                            >
                              {t('database.deploy')}
                            </button>
                          </Show>

                          <Show when={db.status === 'failed'}>
                            <button
                              class="btn btn-sm btn-error"
                              onClick={() => handleDeploy(db)}
                              disabled={deployDatabaseMutation.isPending}
                            >
                              {t('database.redeploy')}
                            </button>
                          </Show>

                          <Show when={db.status === 'stopped'}>
                            <button
                              class="btn btn-sm btn-success"
                              onClick={() => handleStart(db)}
                              disabled={startDatabaseMutation.isPending}
                            >
                              {t('database.start')}
                            </button>
                          </Show>

                          <Show when={db.status === 'running'}>
                            <button
                              class="btn btn-sm btn-warning"
                              onClick={() => handleStop(db)}
                              disabled={stopDatabaseMutation.isPending}
                            >
                              {t('database.stop')}
                            </button>
                            <button
                              class="btn btn-sm btn-info"
                              onClick={() => handleRestart(db)}
                              disabled={restartDatabaseMutation.isPending}
                            >
                              {t('database.restart')}
                            </button>
                          </Show>

                          <button
                            class="btn btn-sm btn-ghost"
                            onClick={() => showConnectionInfo(db)}
                          >
                            {t('database.connection_info')}
                          </button>

                          <button
                            class="btn btn-sm btn-error"
                            onClick={() => openDeleteModal(db)}
                          >
                            {t('database.delete')}
                          </button>
                        </div>
                      </td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </Show>
      </Show>

      {/* Create Database Modal */}
      <Show when={showCreateModal()}>
        <div class="modal modal-open">
          <div class="modal-box max-w-2xl">
            <h3 class="font-bold text-lg mb-4">{t('database.create_title')}</h3>

            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('database.database_name')}</span>
              </label>
              <input
                type="text"
                placeholder={t('database.name_placeholder')}
                class="input input-bordered"
                value={formData().name}
                onInput={(e) => setFormData({ ...formData(), name: e.currentTarget.value })}
              />
            </div>

            <div class="form-control mt-4">
              <label class="label">
                <span class="label-text">{t('database.database_type')}</span>
              </label>
              <select
                class="select select-bordered"
                value={formData().type}
                onChange={(e) =>
                  setFormData({
                    ...formData(),
                    type: e.currentTarget.value as 'postgresql',
                  })
                }
              >
                <option value="postgresql">{t('database.type_postgresql')}</option>
              </select>
            </div>

            <div class="grid grid-cols-2 gap-4 mt-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">{t('database.version')}</span>
                </label>
                <Show when={!useCustomImage()} fallback={
                  <input
                    type="text"
                    placeholder="例如: docker.io/postgres:16-alpine"
                    class="input input-bordered"
                    value={formData().custom_image}
                    onInput={(e) => setFormData({ ...formData(), custom_image: e.currentTarget.value })}
                    disabled={!useCustomImage()}
                  />
                }>
                  <select
                    class="select select-bordered"
                    value={formData().version}
                    onChange={(e) => setFormData({ ...formData(), version: e.currentTarget.value })}
                  >
                    <option value="17-alpine">PostgreSQL 17 (Alpine)</option>
                    <option value="16-alpine">PostgreSQL 16 (Alpine)</option>
                    <option value="15-alpine">PostgreSQL 15 (Alpine)</option>
                    <option value="14-alpine">PostgreSQL 14 (Alpine)</option>
                    <option value="13-alpine">PostgreSQL 13 (Alpine)</option>
                    <option value="17">PostgreSQL 17</option>
                    <option value="16">PostgreSQL 16</option>
                    <option value="15">PostgreSQL 15</option>
                    <option value="alpine">PostgreSQL Latest (Alpine)</option>
                    <option value="latest">PostgreSQL Latest</option>
                  </select>
                </Show>
                <label class="label cursor-pointer justify-start gap-2 mt-1">
                  <input
                    type="checkbox"
                    class="checkbox checkbox-sm"
                    checked={useCustomImage()}
                    onChange={(e) => setUseCustomImage(e.currentTarget.checked)}
                  />
                  <span class="label-text-alt">使用自定义镜像</span>
                </label>
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text">{t('database.port')}</span>
                </label>
                <div class="flex gap-2">
                  <input
                    type="number"
                    placeholder={t('database.port_placeholder')}
                    class="input input-bordered flex-1"
                    value={formData().port}
                    onInput={(e) =>
                      setFormData({ ...formData(), port: parseInt(e.currentTarget.value) })
                    }
                  />
                  <button class="btn btn-outline" onClick={() => setFormData({ ...formData(), port: generatePort() })}>
                    生成随机端口
                  </button>
                </div>
              </div>
            </div>

            <div class="form-control mt-4">
              <label class="label">
                <span class="label-text">{t('database.username')}</span>
              </label>
              <input
                type="text"
                placeholder={t('database.username_placeholder')}
                class="input input-bordered"
                value={formData().username}
                onInput={(e) => setFormData({ ...formData(), username: e.currentTarget.value })}
              />
            </div>

            <div class="form-control mt-4">
              <label class="label">
                <span class="label-text">{t('database.password')}</span>
              </label>
              <div class="flex gap-2">
                <input
                  type="password"
                  placeholder={t('database.password_placeholder')}
                  class="input input-bordered flex-1"
                  value={formData().password}
                  onInput={(e) => setFormData({ ...formData(), password: e.currentTarget.value })}
                />
                <button class="btn btn-outline" onClick={generatePassword}>
                  {t('database.generate_password')}
                </button>
              </div>
            </div>

            <div class="form-control mt-4">
              <label class="label">
                <span class="label-text">{t('database.database_name_field')}</span>
              </label>
              <input
                type="text"
                placeholder={t('database.database_name_placeholder')}
                class="input input-bordered"
                value={formData().database_name}
                onInput={(e) =>
                  setFormData({ ...formData(), database_name: e.currentTarget.value })
                }
              />
            </div>

            <div class="form-control mt-4">
              <label class="label">
                <span class="label-text">{t('database.data_path')}</span>
              </label>
              <input
                type="text"
                placeholder={t('database.data_path_placeholder')}
                class="input input-bordered"
                value={formData().data_path}
                onInput={(e) => setFormData({ ...formData(), data_path: e.currentTarget.value })}
              />
            </div>

            <div class="modal-action">
              <button class="btn" onClick={() => setShowCreateModal(false)}>
                {t('common.cancel')}
              </button>
              <button
                class="btn btn-primary"
                onClick={handleCreateDatabase}
                disabled={createDatabaseMutation.isPending}
              >
                {createDatabaseMutation.isPending ? (
                  <span class="loading loading-spinner"></span>
                ) : (
                  t('common.save')
                )}
              </button>
            </div>
          </div>
        </div>
      </Show>

      {/* Delete Confirmation Modal */}
      <Show when={showDeleteModal()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg">{t('database.delete')}</h3>
            <p class="py-4">
              {t('database.delete_confirm').replace('{{name}}', selectedDatabase()?.name || '')}
            </p>
            <div class="modal-action">
              <button class="btn" onClick={() => setShowDeleteModal(false)}>
                {t('common.cancel')}
              </button>
              <button
                class="btn btn-error"
                onClick={handleDeleteDatabase}
                disabled={deleteDatabaseMutation.isPending}
              >
                {deleteDatabaseMutation.isPending ? (
                  <span class="loading loading-spinner"></span>
                ) : (
                  t('database.delete')
                )}
              </button>
            </div>
          </div>
        </div>
      </Show>

      {/* Connection Info Modal */}
      <Show when={showConnectionModal()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg mb-4">{t('database.connection_info_title')}</h3>

            <Show when={connectionInfo()}>
              {(info) => (
                <div class="space-y-3">
                  <div>
                    <label class="label">
                      <span class="label-text font-semibold">{t('database.host')}</span>
                    </label>
                    <input
                      type="text"
                      class="input input-bordered w-full"
                      value={info().host}
                      readonly
                    />
                  </div>

                  <div>
                    <label class="label">
                      <span class="label-text font-semibold">{t('database.port')}</span>
                    </label>
                    <input
                      type="text"
                      class="input input-bordered w-full"
                      value={info().port}
                      readonly
                    />
                  </div>

                  <div>
                    <label class="label">
                      <span class="label-text font-semibold">{t('database.username')}</span>
                    </label>
                    <input
                      type="text"
                      class="input input-bordered w-full"
                      value={info().user}
                      readonly
                    />
                  </div>

                  <Show when={info().password}>
                    <div>
                      <label class="label">
                        <span class="label-text font-semibold">{t('database.password')}</span>
                      </label>
                      <input
                        type="text"
                        class="input input-bordered w-full"
                        value={info().password}
                        readonly
                      />
                    </div>
                  </Show>

                  <div>
                    <label class="label">
                      <span class="label-text font-semibold">
                        {t('database.connection_string')}
                      </span>
                    </label>
                    <div class="flex gap-2">
                      <input
                        type="text"
                        class="input input-bordered flex-1"
                        value={info().connection_string}
                        readonly
                      />
                      <button class="btn btn-square" onClick={copyConnectionString}>
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke-width="1.5"
                          stroke="currentColor"
                          class="w-5 h-5"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184"
                          />
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </Show>

            <div class="modal-action">
              <button class="btn" onClick={() => setShowConnectionModal(false)}>
                {t('common.cancel')}
              </button>
            </div>
          </div>
        </div>
      </Show>
    </div>
  )
}

export default DatabaseManagementPage
