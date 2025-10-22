import { createSignal, Show } from 'solid-js';
import { useApiMutation, useApiQuery } from '../api/apiHooksW.ts';
import { statusEndpoint, setup2FAEndpoint, verify2FAEndpoint, disable2FAEndpoint } from '../api/endpoints/auth';
import { QRCodeSVG } from 'solid-qr-code';

// --- API Response Types ---
interface AuthStatusResponse {
  user: { id: number; username: string };
  two_factor_enabled: boolean;
}

interface TwoFASetupResponse {
  secret: string;
  encrypted_secret: string;
  qr_code_url: string;
}

interface Verify2FAResponse {
  recovery_codes: string[];
}

const TwoFactorAuthSettings = () => {
  // --- State Signals ---
  const [setupInfo, setSetupInfo] = createSignal<TwoFASetupResponse | null>(null);
  const [otp, setOtp] = createSignal('');
  const [recoveryCodes, setRecoveryCodes] = createSignal<string[]>([]);
  const [disablePassword, setDisablePassword] = createSignal('');
  const [disableOtp, setDisableOtp] = createSignal('');

  // --- API Queries & Mutations ---
  const authStatusQuery = useApiQuery<AuthStatusResponse>(
    () => ['authStatus'],
    () => statusEndpoint().url
  );

  const setupMutation = useApiMutation<TwoFASetupResponse, void>(
    setup2FAEndpoint(),
    {
      onSuccess: (data) => {
        setSetupInfo(data);
        (document.getElementById('enable_2fa_modal') as any)?.showModal();
      },
    }
  );

  const verifyMutation = useApiMutation<Verify2FAResponse, { otp: string; encrypted_secret: string }>(
    verify2FAEndpoint(),
    {
      onSuccess: (data) => {
        (document.getElementById('enable_2fa_modal') as any)?.close();
        setRecoveryCodes(data.recovery_codes);
        setSetupInfo(null);
        authStatusQuery.refetch();
      },
    }
  );

  const disableMutation = useApiMutation<void, { password: string; otp: string }>(
    disable2FAEndpoint(),
    {
      onSuccess: () => {
        (document.getElementById('disable_2fa_modal') as any)?.close();
        authStatusQuery.refetch();
        alert('2FA has been successfully disabled.');
      },
      onError: (err) => {
        alert(`Failed to disable 2FA: ${err.message}`);
      }
    }
  );

  // --- Event Handlers ---
  const handleEnable = () => setupMutation.mutate();
  const handleVerify = (e: Event) => {
    e.preventDefault();
    if (!setupInfo() || !otp()) return;
    verifyMutation.mutate({ otp: otp(), encrypted_secret: setupInfo()!.encrypted_secret });
  };
  const handleOpenDisableModal = () => {
    setDisablePassword('');
    setDisableOtp('');
    (document.getElementById('disable_2fa_modal') as any)?.showModal();
  };
  const handleConfirmDisable = (e: Event) => {
    e.preventDefault();
    if (!disablePassword() || !disableOtp()) return;
    disableMutation.mutate({ password: disablePassword(), otp: disableOtp() });
  }

  const handlePrint = () => {
    const printSectionId = 'recovery-codes-printable-section';
    const style = document.createElement('style');
    style.innerHTML = `@media print { body > * { display: none !important; } #${printSectionId}, #${printSectionId} * { display: block !important; visibility: visible !important; } #${printSectionId} { position: absolute !important; left: 0 !important; top: 0 !important; width: 100% !important; } }`;
    document.head.appendChild(style);
    window.onafterprint = () => { document.head.removeChild(style); window.onafterprint = null; };
    window.print();
  };

  // --- Computed Signals ---
  const isLoading = () => authStatusQuery.isFetching || setupMutation.isPending || verifyMutation.isPending || disableMutation.isPending;
  const error = () => authStatusQuery.error || setupMutation.error || verifyMutation.error || disableMutation.error;
  const is2FAEnabled = () => authStatusQuery.data?.two_factor_enabled === true;

  return (
    <div class="card bg-base-100 shadow mt-8">
      <div class="card-body">
        <h2 class="card-title">Two-Factor Authentication (2FA)</h2>

        {/* Main Content Area on Page */}
        <Show when={authStatusQuery.isLoading}><p>Loading 2FA status...</p></Show>

              <Show when={!authStatusQuery.isLoading && is2FAEnabled()}>
                <div class="form-control">
                    <label class="label">
                      <span class="label-text font-semibold text-green-600">Status: Enabled</span>
                    </label>
                    <p class="text-sm mb-4">Two-Factor Authentication is currently active on your account.</p>
                    <button onClick={handleOpenDisableModal} disabled={isLoading()} class="btn btn-error btn-sm max-w-xs">
                      {isLoading() ? 'Loading...' : 'Disable 2FA'}
                    </button>
                  </div>
              </Show>
        
              <Show when={!authStatusQuery.isLoading && !is2FAEnabled()}>
                <div class="form-control">
                    <label class="label">
                      <span class="label-text">Status: Not Enabled</span>
                    </label>
                    <p class="text-sm mb-4">Enhance your account security by enabling 2FA.</p>
                    <button onClick={handleEnable} disabled={isLoading()} class="btn btn-primary btn-sm max-w-xs">
                      {isLoading() ? 'Loading...' : 'Enable 2FA'}
                    </button>
                  </div>
              </Show>
        <Show when={recoveryCodes().length > 0}>
          <div id="recovery-codes-printable-section" class="mt-4 p-4 bg-green-100 border border-green-400 rounded">
            <h4 class="font-bold text-green-800">2FA Enabled Successfully!</h4>
            <p class="text-sm text-red-700 font-semibold my-2">Please save these recovery codes in a safe place. You will not be shown them again.</p>
            <div class="p-2 bg-gray-100 rounded font-mono text-center">
              {recoveryCodes().map(code => <div>{code}</div>)}
            </div>
            <div class="mt-4">
              <button onClick={handlePrint} class="btn btn-ghost btn-sm">Print / Save as PDF</button>
            </div>
          </div>
        </Show>

        <Show when={error()}><p class="mt-4 text-red-500">{error()?.message}</p></Show>

        {/* Enable 2FA Modal */}
        <dialog id="enable_2fa_modal" class="modal">
          <div class="modal-box">
            <form method="dialog">
              <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
            </form>
            <h3 class="font-bold text-lg">Enable Two-Factor Authentication</h3>
            <Show when={setupInfo()}>
              <div class="mt-4">
                <p class="font-semibold">Step 1: Scan QR Code</p>
                <p class="text-sm mb-2">Scan this image with your authenticator app (e.g., Google Authenticator).</p>
                <div class="my-4 p-4 bg-white inline-block rounded-lg"><QRCodeSVG value={setupInfo()!.qr_code_url} /></div>
                <p class="text-sm">Or manually enter this secret: <span class="font-mono p-1 bg-base-200 rounded">{setupInfo()!.secret}</span></p>
                <div class="divider"></div>
                <form onSubmit={handleVerify} class="mt-4">
                  <p class="font-semibold">Step 2: Verify Code</p>
                  <p class="text-sm mb-2">Enter the 6-digit code from your app to complete setup.</p>
                  <div class="form-control">
                    <input type="text" value={otp()} onInput={(e) => setOtp(e.currentTarget.value)} class="input input-bordered w-full max-w-xs" placeholder="123456" maxLength={6} required />
                    <button type="submit" disabled={verifyMutation.isPending} class="btn btn-primary mt-2 max-w-xs">
                      {verifyMutation.isPending ? 'Verifying...' : 'Verify & Enable'}
                    </button>
                  </div>
                </form>
              </div>
            </Show>
          </div>
        </dialog>

        {/* Disable 2FA Modal */}
        <dialog id="disable_2fa_modal" class="modal">
          <div class="modal-box">
            <h3 class="font-bold text-lg">Confirm Disabling 2FA</h3>
            <p class="py-4">For your security, please enter your password and a current 2FA code to disable this feature.</p>
            <form onSubmit={handleConfirmDisable}>
              <div class="form-control w-full mb-4">
                <label class="label"><span class="label-text">Current Password</span></label>
                <input type="password" value={disablePassword()} onInput={(e) => setDisablePassword(e.currentTarget.value)} class="input input-bordered w-full" required />
              </div>
              <div class="form-control w-full mb-4">
                <label class="label"><span class="label-text">6-Digit Authentication Code</span></label>
                <input type="text" value={disableOtp()} onInput={(e) => setDisableOtp(e.currentTarget.value)} class="input input-bordered w-full" required maxLength={6} />
              </div>
              <div class="modal-action">
                <form method="dialog"><button class="btn">Cancel</button></form>
                <button type="submit" class="btn btn-error" disabled={disableMutation.isPending}>
                  {disableMutation.isPending ? 'Disabling...' : 'Confirm & Disable'}
                </button>
              </div>
            </form>
          </div>
        </dialog>
      </div>
    </div>
  );
};

export default TwoFactorAuthSettings;