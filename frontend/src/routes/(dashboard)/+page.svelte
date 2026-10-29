<script lang="ts">
	import { onMount } from 'svelte';
	import { toSameOriginApiUrl } from '$lib/config/api-url';
	import { apiBaseUrl } from '$lib/config/backend';
	import { authStore } from '$lib/modules/base/stores/auth';
	import { fetchCurrentUser } from '$lib/modules/base/utils/auth-api';
	import { Badge } from '$lib/modules/base/ui/badge/index.js';
	import { Button } from '$lib/modules/base/ui/button/index.js';
	import * as Card from '$lib/modules/base/ui/card/index.js';

	let health = $state<boolean | null>(null);
	let sample = $state<string>('…');
	let apiError = $state<string | null>(null);
	let meCheck = $state<string>('not tested');
	let meLoading = $state(false);

	onMount(async () => {
		try {
			const healthRes = await fetch(toSameOriginApiUrl(`${apiBaseUrl()}/utils/health-check`));
			health = healthRes.ok ? await healthRes.json() : false;

			const sampleRes = await fetch(toSameOriginApiUrl(`${apiBaseUrl()}/sample`));
			if (sampleRes.ok) {
				const body = (await sampleRes.json()) as { message?: string };
				sample = body.message ?? 'ok';
			} else {
				sample = `HTTP ${sampleRes.status}`;
			}
		} catch (e) {
			apiError = e instanceof Error ? e.message : 'Request failed';
		}
	});

	async function testAuthenticatedMe() {
		const token = authStore.getToken();
		if (!token) {
			meCheck = 'no token in store';
			return;
		}
		meLoading = true;
		try {
			const user = await fetchCurrentUser(token);
			meCheck = `${user.email}${user.is_superuser ? ' (superuser)' : ''}`;
		} catch (e) {
			meCheck = e instanceof Error ? e.message : 'request failed';
		} finally {
			meLoading = false;
		}
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Dashboard</h1>
		<p class="text-muted-foreground">
			API health checks and session verification. Default superuser:
			<code class="rounded bg-muted px-1.5 py-0.5 text-sm">admin@example.com</code>
		</p>
	</div>

	<div class="grid gap-4 md:grid-cols-2">
		<Card.Root>
			<Card.Header>
				<Card.Title>Authenticated /me</Card.Title>
				<Card.Description>Test GET /base/login/me with stored JWT</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<Button onclick={() => void testAuthenticatedMe()} disabled={meLoading}>
					{meLoading ? 'Calling /me…' : 'Test GET /base/login/me'}
				</Button>
				<p class="font-mono text-sm text-muted-foreground">{meCheck}</p>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Health check</Card.Title>
				<Card.Description>GET /api/v1/utils/health-check/</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="flex items-center gap-2">
					<Badge variant={apiError ? 'destructive' : health ? 'default' : 'secondary'}>
						{apiError ? 'ERR' : health === null ? '…' : health ? '200' : '503'}
					</Badge>
					<span class="font-mono text-sm">
						{apiError ?? (health === null ? 'checking…' : String(health))}
					</span>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Sample module</Card.Title>
				<Card.Description>GET /api/v1/sample/</Card.Description>
			</Card.Header>
			<Card.Content>
				<p class="font-mono text-sm">{sample}</p>
			</Card.Content>
		</Card.Root>
	</div>
</div>
