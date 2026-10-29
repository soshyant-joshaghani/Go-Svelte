<script lang="ts">
	import type { Component } from 'svelte';
	import { page } from '$app/stores';
	import BriefcaseIcon from '@lucide/svelte/icons/briefcase';
	import HomeIcon from '@lucide/svelte/icons/home';
	import UsersIcon from '@lucide/svelte/icons/users';
	import { authStore } from '$lib/modules/base/stores/auth';
	import { APP_NAME } from '$lib/modules/base';
	import * as Sidebar from '$lib/modules/base/ui/sidebar/index.js';

	const baseItems: { title: string; href: string; icon: Component }[] = [
		{ title: 'Dashboard', href: '/', icon: HomeIcon },
		{ title: 'Sample Notes', href: '/sample/notes', icon: BriefcaseIcon }
	];

	let items = $derived(
		$authStore.user?.is_superuser
			? [...baseItems, { title: 'Admin', href: '/admin', icon: UsersIcon }]
			: baseItems
	);
</script>

<Sidebar.Root>
	<Sidebar.Header class="border-b border-sidebar-border p-4">
		<p class="text-xs font-semibold uppercase tracking-widest text-muted-foreground">Dashboard</p>
		<p class="text-lg font-bold">{APP_NAME}</p>
	</Sidebar.Header>
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>Menu</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					{#each items as item (item.href)}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton isActive={$page.url.pathname === item.href}>
								{#snippet child({ props })}
									<a href={item.href} {...props}>
										<item.icon />
										<span>{item.title}</span>
									</a>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer class="border-t border-sidebar-border p-4 text-xs text-muted-foreground">
		Go-Svelte From FoxG
	</Sidebar.Footer>
</Sidebar.Root>
