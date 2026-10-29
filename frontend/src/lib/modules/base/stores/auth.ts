import { browser } from '$app/environment';
import { writable } from 'svelte/store';
import { fetchCurrentUser } from '$lib/modules/base/utils/auth-api';

export type AuthUser = {
	id?: string;
	email: string;
	full_name?: string | null;
	is_active?: boolean;
	is_superuser?: boolean;
};

type AuthState = {
	user: AuthUser | null;
	token: string | null;
	isAuthenticated: boolean;
	isLoading: boolean;
};

const STORAGE_TOKEN = 'authToken';
const STORAGE_USER = 'currentUser';

const emptyState: AuthState = {
	user: null,
	token: null,
	isAuthenticated: false,
	isLoading: true
};

function clearStorage() {
	if (!browser) return;
	localStorage.removeItem(STORAGE_TOKEN);
	localStorage.removeItem(STORAGE_USER);
}

function saveToStorage(token: string, user: AuthUser) {
	if (!browser) return;
	localStorage.setItem(STORAGE_TOKEN, token);
	localStorage.setItem(STORAGE_USER, JSON.stringify(user));
}

function readStoredCredentials(): { token: string; user: AuthUser } | null {
	if (!browser) return null;

	const token = localStorage.getItem(STORAGE_TOKEN);
	const raw = localStorage.getItem(STORAGE_USER);

	if (!token || !raw) return null;

	try {
		return { token, user: JSON.parse(raw) as AuthUser };
	} catch {
		clearStorage();
		return null;
	}
}

const { subscribe, set } = writable<AuthState>(emptyState);

async function hydrate() {
	if (!browser) {
		set({ ...emptyState, isLoading: false });
		return;
	}

	const stored = readStoredCredentials();
	if (!stored) {
		set({ user: null, token: null, isAuthenticated: false, isLoading: false });
		return;
	}

	try {
		const user = await fetchCurrentUser(stored.token);
		saveToStorage(stored.token, user);
		set({ user, token: stored.token, isAuthenticated: true, isLoading: false });
	} catch {
		clearStorage();
		set({ user: null, token: null, isAuthenticated: false, isLoading: false });
	}
}

export const authStore = {
	subscribe,

	initialize() {
		void hydrate();
	},

	login(token: string, user: AuthUser) {
		saveToStorage(token, user);
		set({ user, token, isAuthenticated: true, isLoading: false });
	},

	setUser(user: AuthUser) {
		if (browser) {
			localStorage.setItem(STORAGE_USER, JSON.stringify(user));
		}
		set({ user, token: authStore.getToken(), isAuthenticated: true, isLoading: false });
	},

	logout() {
		clearStorage();
		set({ user: null, token: null, isAuthenticated: false, isLoading: false });
	},

	getToken(): string | null {
		if (!browser) return null;
		return localStorage.getItem(STORAGE_TOKEN);
	}
};

if (browser) {
	authStore.initialize();
}
