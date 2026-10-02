import { apiBaseUrl } from '$lib/config/backend';
import { ApiError } from '$lib/modules/base/utils/api-error';
import { formatApiError } from '$lib/modules/base/utils/auth-api';
import { authFetch } from '$lib/modules/base/utils/auth-fetch';

export type ManagedUser = {
	id: string;
	email: string;
	full_name: string | null;
	is_active: boolean;
	is_superuser: boolean;
};

function parseUser(data: Record<string, unknown>): ManagedUser {
	return {
		id: String(data.id),
		email: String(data.email),
		full_name: (data.full_name as string | null) ?? null,
		is_active: Boolean(data.is_active ?? true),
		is_superuser: Boolean(data.is_superuser ?? false)
	};
}

async function parseApiError(res: Response, fallback: string): Promise<never> {
	const body = await res.json().catch(() => ({}));
	throw new ApiError(formatApiError(body.detail, fallback), res.status);
}

export async function listUsers(token: string): Promise<ManagedUser[]> {
	const res = await authFetch(token, `${apiBaseUrl()}/base/users/admin`);
	if (!res.ok) return parseApiError(res, 'Failed to load users');
	const body = await res.json();
	return (body.data as Record<string, unknown>[]).map(parseUser);
}

export async function createUser(
	token: string,
	data: {
		email: string;
		password: string;
		full_name?: string | null;
		is_active?: boolean;
		is_superuser?: boolean;
	}
): Promise<ManagedUser> {
	const res = await authFetch(token, `${apiBaseUrl()}/base/users/admin`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	if (!res.ok) return parseApiError(res, 'Failed to create user');
	return parseUser(await res.json());
}

export async function updateUser(
	token: string,
	userId: string,
	data: {
		email?: string;
		password?: string;
		full_name?: string | null;
		is_active?: boolean;
		is_superuser?: boolean;
	}
): Promise<ManagedUser> {
	const res = await authFetch(token, `${apiBaseUrl()}/base/users/${userId}/admin`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	if (!res.ok) return parseApiError(res, 'Failed to update user');
	return parseUser(await res.json());
}

export async function deleteUser(token: string, userId: string): Promise<void> {
	const res = await authFetch(token, `${apiBaseUrl()}/base/users/${userId}/admin`, {
		method: 'DELETE'
	});
	if (!res.ok) return parseApiError(res, 'Failed to delete user');
}
