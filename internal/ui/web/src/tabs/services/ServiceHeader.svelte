<script lang="ts">
  import StatusPill from '$components/StatusPill.svelte';
  import DetailButton from '$components/DetailButton.svelte';
  import Icon from '$components/Icon.svelte';
  import ParentSiteBadge from './ParentSiteBadge.svelte';
  import ServiceDependencies from './ServiceDependencies.svelte';
  import { goToTab } from '$stores/route';
  import {
    type Service,
    services as allServices,
    serviceLabel,
    detailLabel,
    isServiceWorker,
    parentSiteDomain,
    serviceAction,
    loadServices
  } from '$stores/services';
  import { adminServiceFor } from '$stores/presetSuggestions';
  import { openDashboard } from '$stores/dashboard';
  import { m } from '../../paraglide/messages.js';

  function localDetailLabel(s: Service): string {
    if (s.queue_site) return m.services_labels_queueWorker();
    if (s.horizon_site) return m.services_labels_horizon();
    if (s.stripe_listener_site) return m.services_labels_stripeListener();
    if (s.schedule_worker_site) return m.services_labels_scheduler();
    if (s.reverb_site) return m.services_labels_reverb();
    if (s.worker_site && s.worker_name) return m.services_labels_worker({ name: s.worker_name });
    return detailLabel(s);
  }

  interface Props {
    svc: Service;
  }
  let { svc }: Props = $props();

  const admin = $derived(adminServiceFor(svc, $allServices));

  async function openAdmin() {
    if (!admin) return;
    if (admin.status !== 'active') {
      await serviceAction(admin.name, 'start');
      await loadServices();
    }
    const latest = $allServices.find((s) => s.name === admin.name) || admin;
    openDashboard(latest);
  }

  const isWorker = $derived(isWorker_());
  function isWorker_(): boolean {
    return isServiceWorker(svc);
  }

  const active = $derived(svc.status === 'active');
  const parent = $derived(parentSiteDomain(svc));

  let busy = $state(false);
  async function run(action: Parameters<typeof serviceAction>[1]) {
    busy = true;
    try {
      await serviceAction(svc.name, action);
    } finally {
      busy = false;
    }
  }

  function openSite(domain: string) {
    goToTab('sites', domain);
  }
</script>

<div
  class="flex flex-wrap items-center justify-between gap-y-2 px-3 sm:px-5 py-4 border-b border-gray-100 dark:border-lerd-border shrink-0"
>
  <div class="flex items-center gap-3">
    <div>
      <div class="flex items-center gap-2">
        <span class="font-semibold text-gray-900 dark:text-white text-base">{localDetailLabel(svc)}</span>
        {#if svc.version && !isWorker}
          <span class="text-xs font-normal tabular-nums text-gray-500 dark:text-gray-400">{svc.version}</span>
        {/if}
        <StatusPill tone={active ? 'ok' : 'muted'} label={svc.status} />
      </div>

      {#if parent}
        <ParentSiteBadge domain={parent} />
      {/if}

      {#if !isWorker && svc.site_domains && svc.site_domains.length > 0}
        <div class="flex flex-wrap gap-1 mt-1">
          {#each svc.site_domains as d (d)}
            <button
              onclick={() => openSite(d)}
              class="inline-flex items-center gap-1.5 text-xs font-medium bg-gray-100 dark:bg-white/5 hover:bg-gray-200 dark:hover:bg-white/10 border border-gray-200 dark:border-lerd-border text-gray-700 dark:text-gray-300 rounded-full px-2 py-0.5 transition-colors"
            >
              <span class="w-1.5 h-1.5 rounded-full shrink-0 bg-gray-400"></span>
              {d}
            </button>
          {/each}
        </div>
      {/if}

      {#if svc.depends_on && svc.depends_on.length > 0}
        <ServiceDependencies names={svc.depends_on} />
      {/if}
    </div>
  </div>

  <div class="flex items-center gap-2">
    {#if !isWorker && !active}
      <DetailButton tone="primary" onclick={() => run('start')} disabled={busy} loading={busy}>{m.common_start()}</DetailButton>
    {:else if active}
      <DetailButton onclick={() => run('stop')} disabled={busy} loading={busy}>{m.common_stop()}</DetailButton>
      {#if !isWorker}
        {#snippet restartIcon()}
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
        {/snippet}
        <DetailButton onclick={() => run('restart')} disabled={busy} loading={busy} title={m.sites_restartContainer()} icon={restartIcon}>{m.common_restart()}</DetailButton>
      {/if}
    {/if}
    {#if !isWorker}
      {#snippet pinIcon()}
        {#if svc.pinned}
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 17v5M9 10.76a2 2 0 0 1-1.11 1.79l-1.78.9A2 2 0 0 0 5 15.24V17h14v-1.76a2 2 0 0 0-1.11-1.79l-1.78-.9A2 2 0 0 1 15 10.76V6h1a2 2 0 0 0 0-4H8a2 2 0 0 0 0 4h1v4.76z"/>
          </svg>
        {:else}
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="17" x2="12" y2="22"/>
            <path d="M9 10.76a2 2 0 0 1-1.11 1.79l-1.78.9A2 2 0 0 0 5 15.24V17h14v-1.76a2 2 0 0 0-1.11-1.79l-1.78-.9A2 2 0 0 1 15 10.76V6h1a2 2 0 0 0 0-4H8a2 2 0 0 0 0 4h1v4.76z"/>
          </svg>
        {/if}
      {/snippet}
      <DetailButton
        tone={svc.pinned ? 'warn' : 'secondary'}
        onclick={() => run(svc.pinned ? 'unpin' : 'pin')}
        disabled={busy}
        icon={pinIcon}
        title={svc.pinned ? m.services_unpinTitle() : m.services_pinTitle()}
      >{svc.pinned ? m.services_pinned() : m.services_pin()}</DetailButton>
    {/if}
    {#if !isWorker && svc.custom && !active}
      {#snippet trashIcon()}
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
        </svg>
      {/snippet}
      <DetailButton tone="danger" onclick={() => run('remove')} disabled={busy} icon={trashIcon} title={m.services_removeCustom()}></DetailButton>
    {/if}
    {#snippet externalIcon()}
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
      </svg>
    {/snippet}
    {#if active && svc.dashboard}
      <DetailButton onclick={() => openDashboard(svc)} icon={externalIcon} title={m.services_dashboard()}>{m.services_dashboard()}</DetailButton>
    {/if}
    {#if admin}
      <DetailButton tone="info" onclick={openAdmin} icon={externalIcon} title={m.services_openAdmin({ name: serviceLabel(admin.name) })}>
        {m.services_openAdmin({ name: serviceLabel(admin.name) })}
      </DetailButton>
    {/if}
    {#if !admin && !svc.dashboard && svc.connection_url && active}
      <DetailButton href={svc.connection_url} icon={externalIcon} title={svc.connection_url}>{m.services_openConnection()}</DetailButton>
    {/if}
  </div>
</div>
