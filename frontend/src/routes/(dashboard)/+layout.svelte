<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/modules/base/stores/auth';
	import AppSidebar from '$lib/modules/base/AppSidebar.svelte';
	import Header from '$lib/modules/base/Header.svelte';
	import * as Sidebar from '$lib/modules/base/ui/sidebar/index.js';

	let { children } = $props();

	onMount(() => {
		if (!$authStore.isLoading && !$authStore.isAuthenticated) {
			goto('/login');
		}
	});

	$effect(() => {
		if (!$authStore.isLoading && !$authStore.isAuthenticated) {
			goto('/login');
		}
	});
</script>

{#if $authStore.isLoading}
	<div class="flex min-h-screen items-center justify-center text-muted-foreground">
		Restoring session…
	</div>
{:else if $authStore.isAuthenticated}
	<Sidebar.Provider>
		<AppSidebar />
		<Sidebar.Inset>
			<Header />
			<main class="flex-1 p-6">{@render children()}</main>
		</Sidebar.Inset>
	</Sidebar.Provider>
{/if}
