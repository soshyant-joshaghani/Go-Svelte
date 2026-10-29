<script lang="ts">
	import MoonIcon from '@lucide/svelte/icons/moon';
	import SunIcon from '@lucide/svelte/icons/sun';
	import { Button } from '$lib/modules/base/ui/button/index.js';
	import { themeStore } from '$lib/modules/base/stores/theme';
</script>

<Button
	variant="ghost"
	size="icon"
	class="relative"
	onclick={() => themeStore.toggle()}
	aria-label="Toggle theme"
>
	<SunIcon class="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
	<MoonIcon class="absolute h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
</Button>
