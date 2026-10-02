<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { authStore } from '$lib/modules/base/stores/auth';
	import { themeStore } from '$lib/modules/base/stores/theme';
	import PwaRegister from '$lib/modules/base/PwaRegister.svelte';
	import { APP_NAME } from '$lib/modules/base';

	let { children } = $props();

	onMount(() => {
		authStore.initialize();
		themeStore.initialize();
	});
</script>

<PwaRegister />

<svelte:head>
	<title>{APP_NAME}</title>
	<link rel="icon" href="/favicon.ico" sizes="any" />
</svelte:head>

{@render children()}
