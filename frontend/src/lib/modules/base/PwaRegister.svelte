<script lang="ts">
	import { onMount } from 'svelte';
	import { pwaInfo } from 'virtual:pwa-info';

	type BeforeInstallPromptEvent = Event & {
		prompt: () => Promise<void>;
		userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
	};

	const webManifest = $derived(pwaInfo?.webManifest.linkTag ?? '');

	let deferredPrompt = $state<BeforeInstallPromptEvent | null>(null);
	let canInstall = $state(false);
	let isStandalone = $state(false);
	let isIos = $state(false);
	let showIosHint = $state(false);

	onMount(() => {
		const nav = window.navigator as Navigator & { standalone?: boolean };
		isStandalone =
			window.matchMedia('(display-mode: standalone)').matches || nav.standalone === true;
		isIos = /iphone|ipad|ipod/i.test(window.navigator.userAgent);

		if (pwaInfo && 'serviceWorker' in navigator) {
			void (async () => {
				try {
					const reg = await navigator.serviceWorker.register('/sw.js', { scope: '/' });
					await reg.update().catch(() => undefined);
				} catch (error) {
					console.error('PWA service worker registration failed', error);
				}
			})();
		}

		const onBip = (event: Event) => {
			event.preventDefault();
			deferredPrompt = event as BeforeInstallPromptEvent;
			canInstall = true;
		};
		const onInstalled = () => {
			canInstall = false;
			deferredPrompt = null;
			showIosHint = false;
		};

		window.addEventListener('beforeinstallprompt', onBip);
		window.addEventListener('appinstalled', onInstalled);

		return () => {
			window.removeEventListener('beforeinstallprompt', onBip);
			window.removeEventListener('appinstalled', onInstalled);
		};
	});

	async function installApp() {
		if (!deferredPrompt) return;
		await deferredPrompt.prompt();
		await deferredPrompt.userChoice.catch(() => undefined);
		deferredPrompt = null;
		canInstall = false;
	}
</script>

<svelte:head>
	{#if webManifest}
		{@html webManifest}
	{/if}
	<link rel="apple-touch-icon" href="/apple-touch-icon.png" />
	<meta name="theme-color" content="#09090b" />
	<meta name="mobile-web-app-capable" content="yes" />
	<meta name="apple-mobile-web-app-capable" content="yes" />
	<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
</svelte:head>

{#if !isStandalone && (canInstall || isIos)}
	<div
		class="pointer-events-none fixed inset-x-0 bottom-0 z-[80] flex justify-center p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]"
	>
		<div
			class="pointer-events-auto flex max-w-md items-center gap-3 rounded-2xl border border-white/15 bg-slate-950/90 px-4 py-3 text-sm text-slate-100 shadow-lg backdrop-blur-md"
		>
			{#if canInstall}
				<p class="min-w-0 flex-1 leading-snug">Install Go-Svelte on your device</p>
				<button
					type="button"
					class="shrink-0 rounded-xl bg-sky-500 px-3 py-1.5 font-medium text-slate-950 transition hover:bg-sky-400"
					onclick={installApp}
				>
					Install
				</button>
			{:else if isIos}
				{#if showIosHint}
					<p class="min-w-0 flex-1 leading-snug">
						Tap Share, then <strong>Add to Home Screen</strong>
					</p>
					<button
						type="button"
						class="shrink-0 text-slate-400 hover:text-white"
						onclick={() => (showIosHint = false)}
						aria-label="Dismiss"
					>
						✕
					</button>
				{:else}
					<button
						type="button"
						class="font-medium text-sky-300 hover:text-sky-200"
						onclick={() => (showIosHint = true)}
					>
						Add to Home Screen
					</button>
				{/if}
			{/if}
		</div>
	</div>
{/if}
