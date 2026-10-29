<script lang="ts">
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import { goto } from '$app/navigation';
	import * as Avatar from '$lib/modules/base/ui/avatar/index.js';
	import { Button } from '$lib/modules/base/ui/button/index.js';
	import { authStore } from '$lib/modules/base/stores/auth';

	function handleLogout() {
		authStore.logout();
		goto('/login');
	}
</script>

{#if $authStore.user}
	<div class="flex items-center gap-3">
		<div class="hidden items-center gap-2 sm:flex">
			<Avatar.Root class="h-8 w-8">
				<Avatar.Fallback>{$authStore.user.email.slice(0, 2).toUpperCase()}</Avatar.Fallback>
			</Avatar.Root>
			<div class="text-right text-sm leading-tight">
				<p class="font-medium">{$authStore.user.email}</p>
				<p class="text-xs text-muted-foreground">
					{$authStore.user.is_superuser ? 'SuperAdmin' : 'User'}
				</p>
			</div>
		</div>
		<Button variant="outline" size="sm" onclick={handleLogout}>
			<LogOutIcon class="h-4 w-4" />
			Log out
		</Button>
	</div>
{/if}
