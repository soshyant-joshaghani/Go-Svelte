import { browser } from '$app/environment';
import { writable } from 'svelte/store';

export type Theme = 'light' | 'dark';

const STORAGE_KEY = 'theme';

function getInitial(): Theme {
	if (!browser) return 'dark';
	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored === 'light' || stored === 'dark') return stored;
	return 'dark';
}

function applyTheme(theme: Theme) {
	if (!browser) return;
	document.documentElement.classList.toggle('dark', theme === 'dark');
	localStorage.setItem(STORAGE_KEY, theme);
}

const { subscribe, set, update } = writable<Theme>(getInitial());

export const themeStore = {
	subscribe,
	initialize() {
		if (!browser) return;
		const theme = getInitial();
		applyTheme(theme);
		set(theme);
	},
	toggle() {
		update((current) => {
			const next: Theme = current === 'dark' ? 'light' : 'dark';
			applyTheme(next);
			return next;
		});
	}
};

if (browser) {
	themeStore.initialize();
}
