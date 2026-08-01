# Frontend

SvelteKit 5 frontend compiled as a static SPA.

```bash
npm ci
npm run check
npm run build
```

Production uses `@sveltejs/adapter-static`. The build output is copied directly
into Nginx; no Node process or runtime dependencies are installed in the final
image. API and WebSocket calls use same-origin `/api` routes.
