import { toSameOriginApiUrl } from '$lib/config/api-url';

export async function authFetch(
	token: string,
	url: string,
	init: RequestInit = {}
): Promise<Response> {
	const headers = new Headers(init.headers);
	headers.set('Authorization', `Bearer ${token}`);
	return fetch(toSameOriginApiUrl(url), { ...init, headers });
}
