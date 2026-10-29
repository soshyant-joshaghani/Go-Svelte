<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/modules/base/stores/auth';
	import Authentication from '$lib/modules/base/Authentication.svelte';
	import { APP_NAME } from '$lib/modules/base';

	onMount(() => {
		if ($authStore.isAuthenticated) goto('/');
	});

	$effect(() => {
		if (!$authStore.isLoading && $authStore.isAuthenticated) goto('/');
	});
</script>

<div class="flex min-h-screen flex-col items-center justify-center bg-background p-6">
	<div class="mb-8 text-center">
		<p class="text-xs font-semibold uppercase tracking-widest text-muted-foreground">Welcome to</p>
		<h1 class="text-3xl font-bold">{APP_NAME}</h1>
	</div>
	<div class="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg">
		<Authentication redirectTo="/" />
	</div>
</div>
