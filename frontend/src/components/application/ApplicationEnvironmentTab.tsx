import { Component, createSignal, Show, For, createEffect } from 'solid-js'
import { useQueryClient } from '@tanstack/solid-query'
import { toast } from 'solid-toast'
import { useApiQuery, useApiMutation } from '../../api/apiHooksW.ts'
import { 
  getEnvironmentVariablesEndpoint, 
  createEnvironmentVariableEndpoint, 
  updateEnvironmentVariableEndpoint, 
  deleteEnvironmentVariableEndpoint 
} from '../../api/endpoints'
import type { EnvironmentVariable, CreateEnvironmentVariableRequest } from '../../types/project'

interface ApplicationEnvironmentTabProps {
  applicationUid: string
}

interface VariableFormData {
  key: string
  value: string
  isEncrypted: boolean
}

const ApplicationEnvironmentTab: Component<ApplicationEnvironmentTabProps> = (props) => {
  const queryClient = useQueryClient()

  // State for managing variables
  const [isAddingVariable, setIsAddingVariable] = createSignal(false)
  const [newVariable, setNewVariable] = createSignal<VariableFormData>({ key: '', value: '', isEncrypted: false })
  const [editingVariable, setEditingVariable] = createSignal<string | null>(null)
  const [editingValue, setEditingValue] = createSignal<VariableFormData>({ key: '', value: '', isEncrypted: false })

  // Query environment variables directly for the application
  const environmentVariablesQuery = useApiQuery<EnvironmentVariable[]>(
    () => ['applications', props.applicationUid, 'environment-variables'],
    () => getEnvironmentVariablesEndpoint(props.applicationUid).url,
    { enabled: () => !!props.applicationUid }
  )

  const environmentVariables = () => environmentVariablesQuery.data || []
  const isLoading = () => environmentVariablesQuery.isPending

  const refreshData = async () => {
    await queryClient.invalidateQueries({ queryKey: ['applications', props.applicationUid, 'environment-variables'] })
  }

  // Mutations
  const createVariableMutation = useApiMutation<unknown, CreateEnvironmentVariableRequest>(
    createEnvironmentVariableEndpoint(props.applicationUid),
    {
      onSuccess: () => {
        setNewVariable({ key: '', value: '', isEncrypted: false })
        setIsAddingVariable(false)
        toast.success('Environment variable created successfully')
        void refreshData()
      },
    }
  )

  const updateVariableMutation = useApiMutation<unknown, { uid: string } & CreateEnvironmentVariableRequest>(
    (variables: { uid: string } & CreateEnvironmentVariableRequest) => updateEnvironmentVariableEndpoint(variables.uid),
    {
      body: (variables: { uid: string } & CreateEnvironmentVariableRequest) => ({ 
        key: variables.key, 
        value: variables.value, 
        isEncrypted: variables.isEncrypted 
      }),
      onSuccess: () => {
        setEditingVariable(null)
        setEditingValue({ key: '', value: '', isEncrypted: false })
        toast.success('Environment variable updated successfully')
        void refreshData()
      },
    }
  )

  const deleteVariableMutation = useApiMutation<unknown, { uid: string }>(
    (variables: { uid: string }) => deleteEnvironmentVariableEndpoint(variables.uid),
    {
      onSuccess: () => {
        toast.success('Environment variable deleted successfully')
        void refreshData()
      },
    }
  )

  const isMutating = () => createVariableMutation.isPending || updateVariableMutation.isPending || deleteVariableMutation.isPending

  // Event handlers
  function startAddingNormalVariable() {
    setNewVariable({ key: '', value: '', isEncrypted: false })
    setIsAddingVariable(true)
  }

  function startAddingEncryptedVariable() {
    setNewVariable({ key: '', value: '', isEncrypted: true })
    setIsAddingVariable(true)
  }

  function cancelAdding() {
    setNewVariable({ key: '', value: '', isEncrypted: false })
    setIsAddingVariable(false)
  }

  function createVariable() {
    if (!newVariable().key.trim()) return
    createVariableMutation.mutate(newVariable())
  }

  function startEditing(variable: EnvironmentVariable) {
    setEditingVariable(variable.uid)
    setEditingValue({ 
      key: variable.key, 
      value: variable.value, 
      isEncrypted: variable.isEncrypted 
    })
  }

  function cancelEditing() {
    setEditingVariable(null)
    setEditingValue({ key: '', value: '', isEncrypted: false })
  }

  function updateVariable(uid: string) {
    if (!editingValue().key.trim()) return
    updateVariableMutation.mutate({ uid, ...editingValue() })
  }

  function deleteVariable(uid: string) {
    if (!confirm('Are you sure you want to delete this environment variable?')) return
    deleteVariableMutation.mutate({ uid })
  }

  return (
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h3 class="text-lg font-semibold">Environment Variables</h3>
      </div>

      {/* Environment Variables */}
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <div class="flex justify-between items-center">
            <h4 class="card-title">Environment Variables ({environmentVariables().length})</h4>
            <div class="space-x-2">
              <button
                onClick={startAddingNormalVariable}
                disabled={isMutating() || isAddingVariable()}
                class="btn btn-outline btn-sm"
              >
                Add Normal Variable
              </button>
              <button
                onClick={startAddingEncryptedVariable}
                disabled={isMutating() || isAddingVariable()}
                class="btn btn-primary btn-sm"
              >
                Add Encrypted Variable
              </button>
            </div>
          </div>

          {/* Add Variable Form */}
          <Show when={isAddingVariable()}>
            <div class="border rounded p-4 bg-base-200">
              <h5 class="font-medium mb-3">
                Add {newVariable().isEncrypted ? 'Encrypted' : 'Normal'} Variable
              </h5>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="label">
                    <span class="label-text">Key</span>
                  </label>
                  <input
                    type="text"
                    placeholder="VARIABLE_NAME"
                    value={newVariable().key}
                    onInput={(e) => setNewVariable(prev => ({ ...prev, key: e.currentTarget.value }))}
                    class="input input-bordered w-full"
                    disabled={isMutating()}
                  />
                </div>
                <div>
                  <label class="label">
                    <span class="label-text">Value</span>
                  </label>
                  <input
                    type={newVariable().isEncrypted ? "password" : "text"}
                    placeholder="Variable value"
                    value={newVariable().value}
                    onInput={(e) => setNewVariable(prev => ({ ...prev, value: e.currentTarget.value }))}
                    class="input input-bordered w-full"
                    disabled={isMutating()}
                  />
                </div>
              </div>
              <div class="flex justify-end space-x-2 mt-4">
                <button
                  onClick={cancelAdding}
                  disabled={isMutating()}
                  class="btn btn-ghost btn-sm"
                >
                  Cancel
                </button>
                <button
                  onClick={createVariable}
                  disabled={isMutating() || !newVariable().key.trim()}
                  class="btn btn-primary btn-sm"
                >
                  Add Variable
                </button>
              </div>
            </div>
          </Show>

          {/* Variables List */}
          {isLoading() && (
            <div class="flex justify-center p-6">
              <span class="loading loading-spinner loading-lg"></span>
            </div>
          )}

          <Show when={!isLoading() && environmentVariables().length > 0}>
            <div class="overflow-x-auto">
              <table class="table">
                <thead>
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                    <th>Type</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <For each={environmentVariables()}>
                    {(variable) => (
                      <tr>
                        <td>
                          {editingVariable() === variable.uid ? (
                            <input
                              type="text"
                              value={editingValue().key}
                              onInput={(e) => setEditingValue(prev => ({ ...prev, key: e.currentTarget.value }))}
                              class="input input-bordered input-sm w-full"
                              disabled={isMutating()}
                            />
                          ) : (
                            <span class="font-mono">{variable.key}</span>
                          )}
                        </td>
                        <td>
                          {editingVariable() === variable.uid ? (
                            <input
                              type={editingValue().isEncrypted ? "password" : "text"}
                              value={editingValue().value}
                              onInput={(e) => setEditingValue(prev => ({ ...prev, value: e.currentTarget.value }))}
                              class="input input-bordered input-sm w-full"
                              disabled={isMutating()}
                            />
                          ) : (
                            <span class="font-mono">
                              {variable.isEncrypted ? '*'.repeat(8) : variable.value}
                            </span>
                          )}
                        </td>
                        <td>
                          {editingVariable() === variable.uid ? (
                            <div class="form-control">
                              <label class="label cursor-pointer">
                                <input
                                  type="checkbox"
                                  checked={editingValue().isEncrypted}
                                  onChange={(e) => setEditingValue(prev => ({ ...prev, isEncrypted: e.currentTarget.checked }))}
                                  class="checkbox checkbox-primary"
                                  disabled={isMutating()}
                                />
                                <span class="label-text ml-2">Encrypted</span>
                              </label>
                            </div>
                          ) : (
                            <span class={`badge ${variable.isEncrypted ? 'badge-warning' : 'badge-ghost'}`}>
                              {variable.isEncrypted ? 'Encrypted' : 'Normal'}
                            </span>
                          )}
                        </td>
                        <td>
                          <div class="flex items-center gap-2">
                            {editingVariable() === variable.uid ? (
                              <>
                                <button
                                  class="btn btn-sm btn-success"
                                  onClick={() => updateVariable(variable.uid)}
                                  disabled={isMutating() || !editingValue().key.trim()}
                                >
                                  Save
                                </button>
                                <button
                                  class="btn btn-sm btn-ghost"
                                  onClick={cancelEditing}
                                  disabled={isMutating()}
                                >
                                  Cancel
                                </button>
                              </>
                            ) : (
                              <>
                                <button
                                  class="btn btn-sm btn-ghost"
                                  onClick={() => startEditing(variable)}
                                  disabled={isMutating()}
                                >
                                  Edit
                                </button>
                                <button
                                  class="btn btn-sm btn-ghost text-error"
                                  onClick={() => deleteVariable(variable.uid)}
                                  disabled={isMutating()}
                                >
                                  Delete
                                </button>
                              </>
                            )}
                          </div>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </Show>

          <Show when={!isLoading() && environmentVariables().length === 0}>
            <div class="text-center p-6 text-base-content/70">
              No environment variables found. Click "Add Normal Variable" or "Add Encrypted Variable" to get started.
            </div>
          </Show>
        </div>
      </div>
    </div>
  )
}

export default ApplicationEnvironmentTab