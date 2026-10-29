<script lang="ts">
	import * as Sidebar from '$lib/modules/base/ui/sidebar/index.js';
	import { Separator } from '$lib/modules/base/ui/separator/index.js';
	import ThemeToggle from '$lib/modules/base/ThemeToggle.svelte';
	import UserNav from '$lib/modules/base/UserNav.svelte';
</script>

<header class="flex h-14 shrink-0 items-center gap-2 border-b px-4">
	<Sidebar.Trigger class="-ml-1" />
	<Separator orientation="vertical" class="mr-2 h-4" />
	<div class="flex flex-1 items-center justify-end gap-2">
		<ThemeToggle />
		<UserNav />
	</div>
</header>
