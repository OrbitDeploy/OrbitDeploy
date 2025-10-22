import { createSignal, Show } from 'solid-js';
import type { Component } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import { useI18n } from '../i18n';
import { useApiMutation } from '../lib/apiHooks';
import { getAuthApiUrl } from '../api/config';
import { useAuth } from '../contexts/AuthContext'; // Correctly import useAuth

// Response types
interface LoginStep1Response {
  two_factor_required?: boolean;
  temp_2fa_token?: string;
  access_token?: string;
}

interface LoginStep2Response {
  access_token: string;
}

const LoginPage: Component = () => {
  const { t } = useI18n();
  const navigate = useNavigate();
  const auth = useAuth(); // Get auth context

  // State for form inputs
  const [username, setUsername] = createSignal('');
  const [password, setPassword] = createSignal('');
  const [otp, setOtp] = createSignal('');

  // State for login flow
  const [loginStep, setLoginStep] = createSignal<'credentials' | 'otp'>('credentials');
  const [temp2FAToken, setTemp2FAToken] = createSignal('');

  const handleLoginSuccess = (token: string) => {
    auth.setAccessToken(token); // Use setAccessToken from context
    // Force a full page reload to ensure a clean state, mimicking a manual refresh.
    window.location.href = '/';
  };

  // Mutation for Step 1: Username & Password
  const loginMutation = useApiMutation<LoginStep1Response, any>(
    () => getAuthApiUrl('login'),
    {
      onSuccess: (data) => {
        if (data.two_factor_required && data.temp_2fa_token) {
          setTemp2FAToken(data.temp_2fa_token);
          setLoginStep('otp');
        } else if (data.access_token) {
          handleLoginSuccess(data.access_token);
        }
      },
    }
  );

  // Mutation for Step 2: OTP
  const otpMutation = useApiMutation<LoginStep2Response, any>(
    () => getAuthApiUrl('login2FA'),
    {
      onSuccess: (data) => {
        if (data.access_token) {
          handleLoginSuccess(data.access_token);
        }
      },
    }
  );

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (loginStep() === 'credentials') {
      loginMutation.mutate({ username: username(), password: password() });
    } else {
      otpMutation.mutate({ temp_2fa_token: temp2FAToken(), otp: otp() });
    }
  };

  const isLoading = () => loginMutation.isPending || otpMutation.isPending;
  const error = () => loginMutation.error || otpMutation.error;

  return (
    <div class="min-h-screen bg-base-200 flex items-center justify-center">
      <div class="card w-96 bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-center justify-center mb-6">{t('login.title')}</h2>
          
          <form onSubmit={handleSubmit}>
            <Show when={loginStep() === 'credentials'}>
              <div class="form-control w-full mb-4">
                <label class="label"><span class="label-text">{t('login.username')}</span></label>
                <input type="text" placeholder={t('login.username_placeholder')} class="input input-bordered w-full" value={username()} onInput={(e) => setUsername(e.currentTarget.value)} disabled={isLoading()} required />
              </div>
              <div class="form-control w-full mb-6">
                <label class="label"><span class="label-text">{t('login.password')}</span></label>
                <input type="password" placeholder={t('login.password_placeholder')} class="input input-bordered w-full" value={password()} onInput={(e) => setPassword(e.currentTarget.value)} disabled={isLoading()} required />
              </div>
            </Show>

            <Show when={loginStep() === 'otp'}>
              <div class="form-control w-full mb-6">
                <label class="label"><span class="label-text">6-Digit Authentication Code</span></label>
                <input type="text" placeholder="123456" class="input input-bordered w-full" value={otp()} onInput={(e) => setOtp(e.currentTarget.value)} disabled={isLoading()} required maxLength={6} />
              </div>
            </Show>
            
            {error() && (
              <div class="alert alert-error mb-4">
                <span>{error()?.message}</span>
              </div>
            )}
            
            <div class="form-control">
              <button type="submit" class="btn btn-primary" disabled={isLoading()}>
                {isLoading() && <span class="loading loading-spinner"></span>}
                {isLoading() ? t('login.logging_in') : (loginStep() === 'credentials' ? t('login.login_button') : 'Verify')}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
