import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	kit: {
		adapter: adapter({
			fallback: '200.html',
			precompress: true,
			strict: true
		}),
		csp: {
			mode: 'hash',
			directives: {
				'default-src': ['self'],
				'base-uri': ['self'],
				'connect-src': ['self'],
				'font-src': ['self'],
				'form-action': ['self'],
				'img-src': ['self', 'data:'],
				'object-src': ['none'],
				'script-src': ['self'],
				'style-src': ['self', 'unsafe-inline'],
				'upgrade-insecure-requests': true
			}
		}
	}
};

export default config;
