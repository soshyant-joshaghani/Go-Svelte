import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';
import path from 'node:path';
import { defineConfig } from 'vite';

export default defineConfig({
	envDir: path.resolve(__dirname, '..'),
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			registerType: 'autoUpdate',
			includeAssets: [
				'favicon.ico',
				'pwa-192.png',
				'pwa-512.png',
				'pwa-512-maskable.png',
				'apple-touch-icon.png'
			],
			manifest: {
				id: '/',
				name: 'Go-Svelte',
				short_name: 'Go-Svelte',
				description: 'Go-Svelte dashboard foundation',
				theme_color: '#09090b',
				background_color: '#09090b',
				display: 'standalone',
				orientation: 'any',
				scope: '/',
				start_url: '/',
				lang: 'en',
				icons: [
					{ src: 'pwa-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
					{ src: 'pwa-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
					{
						src: 'pwa-512-maskable.png',
						sizes: '512x512',
						type: 'image/png',
						purpose: 'maskable'
					}
				]
			},
			workbox: {
				globPatterns: [
					'client/**/*.{js,css,ico,png,webmanifest}',
					'prerendered/**/*.{html,json}'
				],
				navigateFallback: null,
				runtimeCaching: [
					{
						urlPattern: ({ url, sameOrigin }) =>
							sameOrigin && url.pathname.startsWith('/api'),
						handler: 'NetworkOnly'
					}
				]
			},
			kit: {
				includeVersionFile: true
			}
		})
	],
	server: {
		port: 5000,
		strictPort: true,
		host: true,
		allowedHosts: ['dashboard.localhost', 'localhost'],
		proxy: {
			'/api': {
				target: process.env.API_PROXY_TARGET ?? 'http://localhost:8000',
				changeOrigin: true
			}
		}
	}
});
