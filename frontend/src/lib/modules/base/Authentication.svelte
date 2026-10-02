<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/modules/base/stores/auth';
	import {
		fetchCurrentUser,
		loginWithPassword,
		signupWithPrivateRoute
	} from '$lib/modules/base/utils/auth-api';

	interface Props {
		redirectTo?: string;
		onSuccess?: () => void;
	}

	const { redirectTo, onSuccess }: Props = $props();

	let activeTab = $state<'login' | 'signup'>('login');
	let email = $state('');
	let password = $state('');
	let fullName = $state('');
	let error = $state('');
	let isLoading = $state(false);

	function switchTab(tab: 'login' | 'signup') {
		activeTab = tab;
		error = '';
		email = '';
		password = '';
		fullName = '';
	}

	async function afterAuth(token: string) {
		const user = await fetchCurrentUser(token);
		authStore.login(token, user);
		onSuccess?.();
		if (!onSuccess && redirectTo) {
			await goto(redirectTo);
		}
	}

	async function handleLogin() {
		error = '';
		isLoading = true;
		try {
			const token = await loginWithPassword(email, password);
			await afterAuth(token);
		} catch (e) {
			console.error('Login error:', e);
			error = e instanceof Error ? e.message : 'Unable to connect to server. Please try again.';
		} finally {
			isLoading = false;
		}
	}

	async function handleSignup() {
		error = '';
		isLoading = true;

		if (!email.match(/^[^\s@]+@[^\s@]+\.[^\s@]+$/)) {
			error = 'Invalid email format';
			isLoading = false;
			return;
		}
		if (!password || password.length < 8) {
			error = 'Password must be at least 8 characters long';
			isLoading = false;
			return;
		}

		try {
			await signupWithPrivateRoute(email, password, fullName);
			const token = await loginWithPassword(email, password);
			await afterAuth(token);
		} catch (e) {
			console.error('Signup error:', e);
			error = e instanceof Error ? e.message : 'Unable to connect to server. Please try again.';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="w-full space-y-6">
	<div class="text-center">
		<h2 class="text-2xl font-bold tracking-tight text-slate-100">Welcome back</h2>
		<p class="mt-2 font-mono text-xs text-slate-500">POST /api/v1/private/users/ · dev signup</p>
	</div>

	<div class="flex border-b border-slate-800">
		<button
			type="button"
			class="flex-1 py-3 text-sm font-medium transition {activeTab === 'login'
				? 'border-b-2 border-sky-400/70 bg-sky-900/30 text-slate-100'
				: 'text-slate-500 hover:text-slate-300'}"
			onclick={() => switchTab('login')}
		>
			Login
		</button>
		<button
			type="button"
			class="flex-1 py-3 text-sm font-medium transition {activeTab === 'signup'
				? 'border-b-2 border-emerald-400/70 bg-emerald-900/30 text-slate-100'
				: 'text-slate-500 hover:text-slate-300'}"
			onclick={() => switchTab('signup')}
		>
			Sign up
		</button>
	</div>

	<div class="space-y-4">
		{#if activeTab === 'login'}
			<form
				onsubmit={(e) => {
					e.preventDefault();
					void handleLogin();
				}}
				class="space-y-4"
			>
				<input
					type="email"
					placeholder="Email"
					bind:value={email}
					disabled={isLoading}
					required
					class="w-full rounded-lg border border-slate-700 bg-slate-900 px-4 py-3 text-slate-100 placeholder-slate-500 transition focus:border-sky-500 focus:ring-2 focus:ring-sky-500/20 focus:outline-none disabled:opacity-50"
				/>
				<input
					type="password"
					placeholder="Password"
					bind:value={password}
					disabled={isLoading}
					required
					class="w-full rounded-lg border border-slate-700 bg-slate-900 px-4 py-3 text-slate-100 placeholder-slate-500 transition focus:border-sky-500 focus:ring-2 focus:ring-sky-500/20 focus:outline-none disabled:opacity-50"
				/>
				<button
					type="submit"
					disabled={isLoading}
					class="w-full rounded-lg bg-sky-500 py-3 text-sm font-semibold text-white transition hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{#if isLoading}
						<span class="inline-flex items-center justify-center gap-2">
							<span
								class="size-4 animate-spin rounded-full border-2 border-white/30 border-t-white"
							></span>
							Logging in…
						</span>
					{:else}
						Login
					{/if}
				</button>
				<p class="text-center text-sm text-slate-500">
					No account?
					<button
						type="button"
						onclick={() => switchTab('signup')}
						class="font-medium text-emerald-400 hover:text-emerald-300"
					>
						Sign up
					</button>
				</p>
			</form>
		{:else}
			<form
				onsubmit={(e) => {
					e.preventDefault();
					void handleSignup();
				}}
				class="space-y-4"
			>
				<input
					type="email"
					placeholder="Email"
					bind:value={email}
					disabled={isLoading}
					required
					class="w-full rounded-lg border border-slate-700 bg-slate-900 px-4 py-3 text-slate-100 placeholder-slate-500 transition focus:border-sky-500 focus:ring-2 focus:ring-sky-500/20 focus:outline-none disabled:opacity-50"
				/>
				<input
					type="text"
					placeholder="Full name (optional)"
					bind:value={fullName}
					disabled={isLoading}
					class="w-full rounded-lg border border-slate-700 bg-slate-900 px-4 py-3 text-slate-100 placeholder-slate-500 transition focus:border-sky-500 focus:ring-2 focus:ring-sky-500/20 focus:outline-none disabled:opacity-50"
				/>
				<input
					type="password"
					placeholder="Password (min. 8 characters)"
					bind:value={password}
					disabled={isLoading}
					required
					class="w-full rounded-lg border border-slate-700 bg-slate-900 px-4 py-3 text-slate-100 placeholder-slate-500 transition focus:border-sky-500 focus:ring-2 focus:ring-sky-500/20 focus:outline-none disabled:opacity-50"
				/>
				<button
					type="submit"
					disabled={isLoading}
					class="w-full rounded-lg bg-emerald-500 py-3 text-sm font-semibold text-white transition hover:bg-emerald-400 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{#if isLoading}
						<span class="inline-flex items-center justify-center gap-2">
							<span
								class="size-4 animate-spin rounded-full border-2 border-white/30 border-t-white"
							></span>
							Creating account…
						</span>
					{:else}
						Create account
					{/if}
				</button>
				<p class="text-center text-sm text-slate-500">
					Already registered?
					<button
						type="button"
						onclick={() => switchTab('login')}
						class="font-medium text-sky-400 hover:text-sky-300"
					>
						Login
					</button>
				</p>
			</form>
		{/if}
	</div>

	{#if error}
		<div
			class="rounded-lg border border-red-500/40 bg-red-950/30 px-4 py-3 text-center text-sm text-red-400"
		>
			{error}
		</div>
	{/if}
</div>
