# Frontend architecture

The UI is SvelteKit in client-side SPA mode with static prerendering. Vite and
Node run only during the Docker build. `@sveltejs/adapter-static` writes the
production site to `build/`, which is copied into the Nginx image.

Browser requests use same-origin `/api/v1` paths. WebSocket URLs are derived
from the current page protocol and host (`/api/ws`), so no production Vite
environment variables or cross-origin policy are required.
