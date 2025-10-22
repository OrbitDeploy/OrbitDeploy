import { createSignal } from 'solid-js'
import type { Component } from 'solid-js'
import { toast } from 'solid-toast'
import { useI18n } from '../i18n'
import { useApiMutation } from '../api/apiHooksW.ts'
import { changePasswordEndpoint } from '../api/endpoints/auth'
import TwoFactorAuthSettings from '../components/TwoFactorAuthSettings' // Import the new component

const ChangePasswordPage: Component = () => {
  const { t } = useI18n()
  const [currentPassword, setCurrentPassword] = createSignal('')
  const [newPassword, setNewPassword] = createSignal('')
  const [confirmPassword, setConfirmPassword] = createSignal('')

  // Change password mutation using the unified API pattern
  const changePasswordMutation = useApiMutation<unknown, { current_password: string; new_password: string }>(
    changePasswordEndpoint(),
    {
      onSuccess: () => {
        toast.success(t('change_password.success_message'))
        // Clear form
        setCurrentPassword('')
        setNewPassword('')
        setConfirmPassword('')
      },
      onError: (error) => {
        toast.error(error instanceof Error ? error.message : t('change_password.error_change_failed'))
      },
    }
  )

  const handleSubmit = async (e: Event) => {
    e.preventDefault()
    
    // Validate input
    if (!currentPassword().trim() || !newPassword().trim()) {
      toast.error(t('change_password.error_empty_fields'))
      return
    }
    
    if (newPassword().length < 6) {
      toast.error(t('change_password.error_password_length'))
      return
    }
    
    if (newPassword() !== confirmPassword()) {
      toast.error(t('change_password.error_password_mismatch'))
      return
    }

    // Use the mutation to change password
    changePasswordMutation.mutate({
      current_password: currentPassword(),
      new_password: newPassword(),
    })
  }

  return (
    <div class="container mx-auto p-6">
      {/* Header */}
      <div class="mb-6">
        <h1 class="text-3xl font-bold text-base-content">{t('change_password.title')}</h1>
        <p class="text-base-content/70 mt-2">{t('change_password.description')}</p>
      </div>

      <div class="max-w-md mx-auto">
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <form onSubmit={handleSubmit}>
              <div class="form-control w-full mb-4">
                <label class="label">
                  <span class="label-text">{t('change_password.current_password')}</span>
                </label>
                <input
                  type="password"
                  placeholder={t('change_password.current_password_placeholder')}
                  class="input input-bordered w-full"
                  value={currentPassword()}
                  onInput={(e) => setCurrentPassword(e.target.value)}
                  disabled={changePasswordMutation.isPending}
                  required
                />
              </div>
              
              <div class="form-control w-full mb-4">
                <label class="label">
                  <span class="label-text">{t('change_password.new_password')}</span>
                </label>
                <input
                  type="password"
                  placeholder={t('change_password.new_password_placeholder')}
                  class="input input-bordered w-full"
                  value={newPassword()}
                  onInput={(e) => setNewPassword(e.target.value)}
                  disabled={changePasswordMutation.isPending}
                  required
                />
                <label class="label">
                  <span class="label-text-alt">{t('change_password.password_length_hint')}</span>
                </label>
              </div>
              
              <div class="form-control w-full mb-6">
                <label class="label">
                  <span class="label-text">{t('change_password.confirm_password')}</span>
                </label>
                <input
                  type="password"
                  placeholder={t('change_password.confirm_password_placeholder')}
                  class="input input-bordered w-full"
                  value={confirmPassword()}
                  onInput={(e) => setConfirmPassword(e.target.value)}
                  disabled={changePasswordMutation.isPending}
                  required
                />
              </div>
              
              <div class="form-control">
                <button
                  type="submit"
                  class="btn btn-primary"
                  disabled={changePasswordMutation.isPending}
                >
                  {changePasswordMutation.isPending && <span class="loading loading-spinner"></span>}
                  {changePasswordMutation.isPending ? t('change_password.changing') : t('change_password.change_button')}
                </button>
              </div>
            </form>
          </div>
        </div>

        {/* 2FA Settings Component */}
        <div class="mt-8">
          <TwoFactorAuthSettings />
        </div>

      </div>
    </div>
  )
}

export default ChangePasswordPage