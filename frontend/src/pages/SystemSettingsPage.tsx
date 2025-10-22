import { createSignal, onMount } from 'solid-js';
import type { Component } from 'solid-js';
import { toast } from 'solid-toast';
import { useI18n } from '../i18n';
import { useApiMutation, useApiQuery } from '../lib/apiHooks';
import { getSystemSettingEndpoint, updateSystemSettingEndpoint } from '../api/endpoints';
import EnvironmentCheckModal from '../components/EnvironmentCheckModal';

const SystemSettingsPage: Component = () => {
  const { t } = useI18n();
  const [domain, setDomain] = createSignal('');
  const [isEnvCheckModalOpen, setIsEnvCheckModalOpen] = createSignal(false);

  const endpoint = () => getSystemSettingEndpoint('domain');

  const systemSettingQuery = useApiQuery<{ value: string }>(
    () => ['systemSetting', 'domain'],
    () => endpoint().url,
    {
      onSuccess: (data) => {
        if (data) {
          setDomain(data.value);
        }
      },
    }
  );

  const updateSystemSettingMutation = useApiMutation<unknown, { value: string }>(
    () => updateSystemSettingEndpoint('domain').url,
    {
      method: 'PUT',
      onSuccess: () => {
        toast.success(t('system_settings.success_message'));
      },
      onError: (error: any) => {
        toast.error(error.message || t('system_settings.error_update_failed'));
      },
    }
  );

  const handleSubmit = async (e: Event) => {
    e.preventDefault();

    if (!domain().trim()) {
      toast.error(t('system_settings.error_empty_domain'));
      return;
    }

    updateSystemSettingMutation.mutate({ value: domain() });
  };

  return (
    <div class="container mx-auto p-6">
      <div class="mb-6">
        <h1 class="text-3xl font-bold text-base-content">{t('system_settings.title')}</h1>
        <p class="text-base-content/70 mt-2">{t('system_settings.description')}</p>
      </div>

      <div class="max-w-md mx-auto">
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <form onSubmit={handleSubmit}>
              <div class="form-control w-full mb-4">
                <label class="label">
                  <span class="label-text">{t('system_settings.domain')}</span>
                </label>
                <input
                  type="text"
                  placeholder={t('system_settings.domain_placeholder')}
                  class="input input-bordered w-full"
                  value={domain()}
                  onInput={(e) => setDomain(e.target.value)}
                  disabled={updateSystemSettingMutation.isPending || systemSettingQuery.isFetching}
                  required
                />
              </div>

              <div class="form-control">
                <button
                  type="submit"
                  class="btn btn-primary"
                  disabled={updateSystemSettingMutation.isPending}
                >
                  {updateSystemSettingMutation.isPending && <span class="loading loading-spinner"></span>}
                  {updateSystemSettingMutation.isPending ? t('system_settings.saving') : t('system_settings.save_button')}
                </button>
              </div>
            </form>
          </div>
        </div>

        <div class="card bg-base-100 shadow mt-6">
          <div class="card-body">
            <h2 class="card-title">{t('system_settings.environment_check')}</h2>
            <p>{t('system_settings.environment_check_description')}</p>
            <div class="card-actions justify-end">
              <button class="btn btn-outline" onClick={() => setIsEnvCheckModalOpen(true)}>{t('system_settings.run_check')}</button>
            </div>
          </div>
        </div>
      </div>

      <EnvironmentCheckModal 
        isOpen={isEnvCheckModalOpen()} 
        onClose={() => setIsEnvCheckModalOpen(false)} 
      />
    </div>
  );
};

export default SystemSettingsPage;
