import { browser } from '$app/environment';
import { toSameOriginApiUrl } from '$lib/config/api-url';
import { apiBaseUrl } from '$lib/config/backend';
import type { AuthUser } from '$lib/modules/base/stores/auth';
import { authFetch } from '$lib/modules/base/utils/auth-fetch';

export function formatApiError(detail: unknown, fallback: string): string {
	if (typeof detail === 'string') return detail;
	if (Array.isArray(detail)) {
		return detail
			.map((item) => {
				if (typeof item === 'object' && item && 'msg' in item) {
					return String((item as { msg: unknown }).msg);
				}
				return String(item);
			})
			.join(', ');
	}
	return fallback;
}

export async function fetchCurrentUser(token: string): Promise<AuthUser> {
	const res = await authFetch(token, `${apiBaseUrl()}/base/login/me`);
	if (!res.ok) throw new Error('Failed to fetch user profile');
	return res.json();
}

export async function loginWithPassword(email: string, password: string): Promise<string> {
	const formData = new URLSearchParams();
	formData.append('username', email);
	formData.append('password', password);

	const response = await fetch(toSameOriginApiUrl(`${apiBaseUrl()}/base/login/access-token`), {
		method: 'POST',
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
		body: formData.toString()
	});

	if (!response.ok) {
		const data = await response.json().catch(() => ({}));
		throw new Error(formatApiError(data.detail, 'Invalid email or password'));
	}

	const { access_token } = await response.json();
	return access_token as string;
}

export async function signupWithPrivateRoute(
	email: string,
	password: string,
	fullName?: string
): Promise<void> {
	if (!browser) return;

	const response = await fetch(toSameOriginApiUrl(`${apiBaseUrl()}/private/users`), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({
			email: email.trim(),
			password,
			...(fullName?.trim() ? { full_name: fullName.trim() } : {})
		})
	});

	if (!response.ok) {
		const data = await response.json().catch(() => ({}));
		throw new Error(formatApiError(data.detail, 'Sign up failed'));
	}
}
