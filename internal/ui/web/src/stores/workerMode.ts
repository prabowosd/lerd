import { writable } from 'svelte/store';
import { apiJson, apiFetch } from '$lib/api';

export type WorkerExecMode = 'exec' | 'container';

export const workerExecMode = writable<WorkerExecMode>('exec');
export const workerModeApplies = writable<boolean>(false);
export const workerModeLoading = writable<boolean>(false);

interface SettingsResponse {
  worker_exec_mode?: string;
  worker_mode_applies?: boolean;
}

function normalizeMode(v: unknown): WorkerExecMode {
  return v === 'container' ? 'container' : 'exec';
}

export async function loadWorkerMode() {
  try {
    const res = await apiJson<SettingsResponse>('/api/settings');
    workerExecMode.set(normalizeMode(res.worker_exec_mode));
    workerModeApplies.set(Boolean(res.worker_mode_applies));
  } catch {
    /* keep previous */
  }
}

export async function setWorkerMode(mode: WorkerExecMode): Promise<boolean> {
  workerModeLoading.set(true);
  try {
    const res = await apiFetch('/api/settings/worker-mode', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ mode })
    });
    const data = (await res.json()) as { ok?: boolean };
    if (res.ok && data.ok) {
      workerExecMode.set(mode);
      return true;
    }
    return false;
  } catch {
    return false;
  } finally {
    workerModeLoading.set(false);
  }
}
