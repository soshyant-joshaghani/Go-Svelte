<script lang="ts">
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/modules/base/stores/auth';
	import { ApiError } from '$lib/modules/base/utils/api-error';
	import {
		createUser,
		deleteUser,
		listUsers,
		updateUser,
		type ManagedUser
	} from '$lib/modules/base/users/api';
	import { Button } from '$lib/modules/base/ui/button/index.js';
	import * as Card from '$lib/modules/base/ui/card/index.js';
	import { Input } from '$lib/modules/base/ui/input/index.js';
	import { Label } from '$lib/modules/base/ui/label/index.js';
	import { Switch } from '$lib/modules/base/ui/switch/index.js';
	import { cn } from '$lib/utils';

	let users = $state<ManagedUser[]>([]);
	let selectedId = $state<string | null>(null);
	let email = $state('');
	let fullName = $state('');
	let password = $state('');
	let isActive = $state(false);
	let isSuperuser = $state(false);
	let status = $state('');
	let loading = $state(false);
	let saving = $state(false);

	$effect(() => {
		if ($authStore.user && !$authStore.user.is_superuser) {
			goto('/');
		}
	});

	$effect(() => {
		if (!$authStore.token || !$authStore.user?.is_superuser) return;
		void refresh();
	});

	async function refresh() {
		const token = $authStore.token;
		if (!token) return;
		loading = true;
		status = '';
		try {
			users = await listUsers(token);
			status = `${users.length} user(s)`;
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				authStore.logout();
				goto('/login');
				return;
			}
			users = [];
			status = e instanceof Error ? e.message : 'Failed to load users';
		} finally {
			loading = false;
		}
	}

	function clearForm() {
		selectedId = null;
		email = '';
		fullName = '';
		password = '';
		isActive = false;
		isSuperuser = false;
		status = 'New user';
	}

	function selectUser(managedUser: ManagedUser) {
		selectedId = managedUser.id;
		email = managedUser.email;
		fullName = managedUser.full_name ?? '';
		password = '';
		isActive = managedUser.is_active;
		isSuperuser = managedUser.is_superuser;
		status = `Editing: ${managedUser.email}`;
	}

	async function handleSave(event: SubmitEvent) {
		event.preventDefault();
		const token = $authStore.token;
		if (!token) return;
		const trimmedEmail = email.trim();
		if (!trimmedEmail) {
			status = 'Email is required';
			return;
		}
		saving = true;
		status = '';
		try {
			if (selectedId) {
				await updateUser(token, selectedId, {
					email: trimmedEmail,
					full_name: fullName.trim() || null,
					password: password || undefined,
					is_active: isActive,
					is_superuser: isSuperuser
				});
				status = 'User updated';
			} else {
				if (password.length < 8) {
					status = 'Password must be at least 8 characters';
					saving = false;
					return;
				}
				await createUser(token, {
					email: trimmedEmail,
					password,
					full_name: fullName.trim() || null,
					is_active: isActive,
					is_superuser: isSuperuser
				});
				status = 'User created';
			}
			clearForm();
			await refresh();
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				authStore.logout();
				goto('/login');
				return;
			}
			status = e instanceof Error ? e.message : 'Save failed';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		const token = $authStore.token;
		if (!token || !selectedId) return;
		if ($authStore.user?.id === selectedId) {
			status = 'Cannot delete your own account here';
			return;
		}
		saving = true;
		status = '';
		try {
			await deleteUser(token, selectedId);
			status = 'User deleted';
			clearForm();
			await refresh();
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				authStore.logout();
				goto('/login');
				return;
			}
			status = e instanceof Error ? e.message : 'Delete failed';
		} finally {
			saving = false;
		}
	}
</script>

{#if $authStore.user?.is_superuser}
	<div class="space-y-6">
		<div class="flex flex-wrap items-start justify-between gap-4">
			<div>
				<h1 class="text-3xl font-bold tracking-tight">Users</h1>
				<p class="text-muted-foreground">Manage user accounts and permissions</p>
			</div>
			<Button type="button" variant="secondary" onclick={clearForm}>Add user</Button>
		</div>

		<div class="grid gap-6 lg:grid-cols-2">
			<Card.Root>
				<Card.Header>
					<Card.Title>All users</Card.Title>
					<Card.Description>{loading ? 'Loading…' : status}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-2">
					{#if !loading && users.length === 0}
						<p class="text-sm text-muted-foreground">No users yet</p>
					{:else}
						{#each users as managedUser (managedUser.id)}
							<Button
								type="button"
								variant={managedUser.id === selectedId ? 'default' : 'outline'}
								class="h-auto w-full justify-start py-2 text-left"
								onclick={() => selectUser(managedUser)}
							>
								<span class="truncate">
									{managedUser.email}{managedUser.is_superuser ? ' (superuser)' : ''}
								</span>
							</Button>
						{/each}
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header>
					<Card.Title>{selectedId ? 'Edit user' : 'New user'}</Card.Title>
					<Card.Description>
						{selectedId ? 'Update account details' : 'Create a new account'}
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<form class="space-y-4" onsubmit={handleSave}>
						<div class="space-y-2">
							<Label for="admin-email">Email</Label>
							<Input id="admin-email" type="email" bind:value={email} required />
						</div>
						<div class="space-y-2">
							<Label for="admin-full-name">Full name</Label>
							<Input id="admin-full-name" type="text" bind:value={fullName} />
						</div>
						<div class="space-y-2">
							<Label for="admin-password">
								Password {selectedId ? '(leave blank to keep)' : '(required)'}
							</Label>
							<Input
								id="admin-password"
								type="password"
								bind:value={password}
								autocomplete="new-password"
							/>
						</div>
						<div class="flex items-center justify-between gap-4">
							<Label for="admin-active">Is active</Label>
							<Switch id="admin-active" bind:checked={isActive} />
						</div>
						<div class="flex items-center justify-between gap-4">
							<Label for="admin-superuser">Is superuser</Label>
							<Switch id="admin-superuser" bind:checked={isSuperuser} />
						</div>
						<div class="flex flex-wrap gap-2 pt-2">
							<Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</Button>
							<Button
								type="button"
								variant="destructive"
								disabled={saving || !selectedId}
								onclick={() => void handleDelete()}
							>
								Delete
							</Button>
						</div>
						{#if status}
							<p class={cn('text-sm', status.includes('failed') && 'text-destructive')}>{status}</p>
						{/if}
					</form>
				</Card.Content>
			</Card.Root>
		</div>
	</div>
{/if}
