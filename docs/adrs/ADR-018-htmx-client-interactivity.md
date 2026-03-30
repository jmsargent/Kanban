# ADR-018: htmx for Client-Side Interactivity

## Status

Accepted

## Context

The kanban-web-view needs client-side interactivity for: loading card details without a full page reload, submitting the add-task form without navigating away from the board, and updating the board after mutations. We need to choose how to implement this interactivity.

The project has zero budget, a single developer, and the server already renders HTML via Go `html/template`. The team has no frontend framework expertise and wants to avoid a JavaScript build toolchain.

## Decision

Use htmx 2.x (BSD-2-Clause) for all client-side interactivity. The server returns HTML fragments for partial page updates. All interactive behavior is declared via HTML attributes (`hx-get`, `hx-post`, `hx-swap`, `hx-trigger`). No custom JavaScript is written.

htmx is loaded as a single script tag, either self-hosted or from a CDN (unpkg.com). Self-hosting is preferred for CSP compliance and offline resilience.

## Alternatives Considered

- **Vanilla JavaScript (fetch + DOM manipulation)**: Zero dependencies. Rejected because it requires writing and maintaining custom JavaScript for every interaction (AJAX calls, DOM updates, error handling). This is more code to write, test, and maintain for a solo developer. The "no explicit JavaScript" benefit of htmx directly addresses the user's preference.

- **Full SPA framework (React, Vue, Svelte)**: Rich client-side interactivity. Rejected because it requires a JavaScript build toolchain (npm, bundler, transpiler), duplicates rendering logic (server templates + client components), adds significant complexity for a 5-page application, and violates the zero-budget/minimal-complexity constraints. Resume-driven development for this scale.

- **Server-rendered full page reloads (no JS)**: Simplest possible approach. Rejected because it degrades UX for card detail views (full page reload to see a single card) and form submissions (user loses board context on add-task). htmx provides the UX improvement with minimal added complexity.

## Consequences

- Positive: No JavaScript build toolchain. No npm. No bundler. No transpiler.
- Positive: Server remains the single source of rendered HTML. No client-side rendering logic to maintain.
- Positive: htmx is 14KB gzipped. Minimal impact on page load time.
- Positive: HTML attributes are declarative and self-documenting. Behavior is visible in the template files.
- Negative: htmx is a runtime dependency. If the CDN is down (mitigated by self-hosting), interactivity breaks. Fallback: forms still work via standard form submission (progressive enhancement).
- Negative: htmx has a learning curve for developers unfamiliar with it. Mitigated by excellent documentation and the limited number of attributes needed (hx-get, hx-post, hx-swap, hx-target, hx-trigger).
