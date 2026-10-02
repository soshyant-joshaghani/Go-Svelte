<script lang="ts">
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/modules/base/stores/auth';
	import { ApiError } from '$lib/modules/base/utils/api-error';
	import {
		createNote,
		deleteNote,
		listNotes,
		updateNote,
		type Note
	} from '$lib/modules/apps/sample/api';
	import { Button } from '$lib/modules/base/ui/button/index.js';
	import * as Card from '$lib/modules/base/ui/card/index.js';
	import { Input } from '$lib/modules/base/ui/input/index.js';
	import { Label } from '$lib/modules/base/ui/label/index.js';
	import * as Sheet from '$lib/modules/base/ui/sheet/index.js';
	import * as Table from '$lib/modules/base/ui/table/index.js';

	let notes = $state<Note[]>([]);
	let title = $state('');
	let content = $state('');
	let error = $state<string | null>(null);
	let loading = $state(false);
	let saving = $state(false);
	let selectedNote = $state<Note | null>(null);
	let editTitle = $state('');
	let editContent = $state('');
	let sheetOpen = $state(false);

	function handleUnauthorized() {
		authStore.logout();
		goto('/login');
	}

	async function refresh() {
		if ($authStore.isLoading || !$authStore.token) return;

		loading = true;
		error = null;
		try {
			notes = await listNotes();
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				handleUnauthorized();
				return;
			}
			error = e instanceof Error ? e.message : 'Failed to load notes';
		} finally {
			loading = false;
		}
	}

	function openEdit(note: Note) {
		selectedNote = note;
		editTitle = note.title;
		editContent = note.content;
		error = null;
		sheetOpen = true;
	}

	function closeEdit() {
		selectedNote = null;
		editTitle = '';
		editContent = '';
		sheetOpen = false;
	}

	async function handleCreate(event: SubmitEvent) {
		event.preventDefault();
		if (!$authStore.token || !title.trim()) return;
		saving = true;
		error = null;
		try {
			await createNote({ title: title.trim(), content: content.trim() });
			title = '';
			content = '';
			await refresh();
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				handleUnauthorized();
				return;
			}
			error = e instanceof Error ? e.message : 'Failed to create note';
		} finally {
			saving = false;
		}
	}

	async function handleSaveEdit(event: SubmitEvent) {
		event.preventDefault();
		if (!$authStore.token || !selectedNote || !editTitle.trim()) return;
		saving = true;
		error = null;
		try {
			await updateNote(selectedNote.id, {
				title: editTitle.trim(),
				content: editContent.trim()
			});
			closeEdit();
			await refresh();
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				handleUnauthorized();
				return;
			}
			error = e instanceof Error ? e.message : 'Failed to update note';
		} finally {
			saving = false;
		}
	}

	async function handleDelete(id: string) {
		if (!$authStore.token) return;
		saving = true;
		error = null;
		try {
			await deleteNote(id);
			if (selectedNote?.id === id) closeEdit();
			await refresh();
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) {
				handleUnauthorized();
				return;
			}
			error = e instanceof Error ? e.message : 'Failed to delete note';
		} finally {
			saving = false;
		}
	}

	$effect(() => {
		if (!$authStore.isLoading && $authStore.token) {
			void refresh();
		}
	});

	const busy = $derived(loading || saving);
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Sample Notes</h1>
		<p class="text-muted-foreground">Canonical CRUD module — Router → Service → Repository</p>
	</div>

	<div class="grid gap-6 lg:grid-cols-3">
		<Card.Root class="lg:col-span-1">
			<Card.Header>
				<Card.Title>New note</Card.Title>
				<Card.Description>Create a note via POST /sample/notes</Card.Description>
			</Card.Header>
			<Card.Content>
				<form class="space-y-4" onsubmit={handleCreate}>
					<div class="space-y-2">
						<Label for="title">Title</Label>
						<Input id="title" bind:value={title} required />
					</div>
					<div class="space-y-2">
						<Label for="content">Content</Label>
						<Input id="content" bind:value={content} />
					</div>
					<Button type="submit" disabled={busy}>Create note</Button>
				</form>
			</Card.Content>
		</Card.Root>

		<Card.Root class="lg:col-span-2">
			<Card.Header>
				<Card.Title>Your notes</Card.Title>
				<Card.Description>{loading ? 'Loading…' : `${notes.length} note(s)`}</Card.Description>
			</Card.Header>
			<Card.Content>
				{#if error}
					<p class="mb-4 text-sm text-destructive">{error}</p>
				{/if}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Title</Table.Head>
							<Table.Head>Content</Table.Head>
							<Table.Head class="w-[80px]"></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#if notes.length === 0}
							<Table.Row>
								<Table.Cell colspan={3} class="text-muted-foreground">No notes yet.</Table.Cell>
							</Table.Row>
						{:else}
							{#each notes as note (note.id)}
								<Table.Row
									class="cursor-pointer"
									data-state={selectedNote?.id === note.id ? 'selected' : undefined}
									onclick={() => openEdit(note)}
								>
									<Table.Cell class="font-medium">{note.title}</Table.Cell>
									<Table.Cell class="text-muted-foreground">{note.content || '—'}</Table.Cell>
									<Table.Cell>
										<Button
											variant="ghost"
											size="icon"
											onclick={(e) => {
												e.stopPropagation();
												void handleDelete(note.id);
											}}
											disabled={busy}
											aria-label="Delete note"
										>
											<Trash2Icon class="h-4 w-4 text-destructive" />
										</Button>
									</Table.Cell>
								</Table.Row>
							{/each}
						{/if}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	</div>

	<Sheet.Root
		bind:open={sheetOpen}
		onOpenChange={(open) => {
			if (!open) closeEdit();
		}}
	>
		<Sheet.Content side="right" class="sm:max-w-md">
			<form class="flex h-full flex-col" onsubmit={handleSaveEdit}>
				<Sheet.Header>
					<Sheet.Title>Edit note</Sheet.Title>
					<Sheet.Description>Update the title and content, then save.</Sheet.Description>
				</Sheet.Header>
				<div class="flex flex-1 flex-col gap-4 overflow-y-auto px-4 py-2">
					<div class="space-y-2">
						<Label for="edit-title">Title</Label>
						<Input id="edit-title" bind:value={editTitle} required />
					</div>
					<div class="space-y-2">
						<Label for="edit-content">Content</Label>
						<Input id="edit-content" bind:value={editContent} />
					</div>
				</div>
				<Sheet.Footer class="flex-row gap-2 sm:justify-between">
					<Button
						type="button"
						variant="destructive"
						disabled={busy || !selectedNote}
						onclick={() => selectedNote && void handleDelete(selectedNote.id)}
					>
						Delete
					</Button>
					<div class="flex gap-2">
						<Button type="button" variant="outline" onclick={closeEdit} disabled={busy}>
							Cancel
						</Button>
						<Button type="submit" disabled={busy}>Save changes</Button>
					</div>
				</Sheet.Footer>
			</form>
		</Sheet.Content>
	</Sheet.Root>
</div>
