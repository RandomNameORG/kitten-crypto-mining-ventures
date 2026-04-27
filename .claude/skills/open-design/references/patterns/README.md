# Patterns — golden-reference implementations

Twelve `.jsx` starter files mirroring `open-codesign`'s upstream design-skills tree. Each contains one or more named variants you can lift as a starting point for the corresponding artifact type.

Every file begins with a `// when_to_use:` comment — read that first to confirm the pattern matches the brief.

## JSX → HTML adapter (important)

These files are React components (`<script type="text/babel">`) because upstream ships an Electron iframe sandbox that compiles JSX on the fly. **Your output is a plain `.html` file**, so you must translate as you read:

| JSX | HTML equivalent |
|---|---|
| `className="foo"` | `class="foo"` |
| `style={{ background: t.accent, padding: 24 }}` | `style="background: var(--color-accent); padding: 24px;"` |
| `{t.accent}` / `TWEAK_DEFAULTS.accent` | CSS custom property `var(--color-accent)` declared on `:root` |
| `const SERIF = "'Fraunces', ...";` | `--font-display: 'Fraunces', ...;` on `:root` |
| `{features.map((f) => (...))}` | Write the markup out, one block per item (artifact is mockup, not templating) |
| React state, hooks, handlers | Drop — artifacts are static surfaces. Exception: `<details>`, CSS `:checked` toggles, `setInterval` dashboard clock. |
| Inline SVGs | Copy verbatim |

**Extract, don't copy-paste.** What you want from a pattern:
- Section hierarchy and rhythm (where sections break, how dense each is).
- Copy tone and specifics (eyebrow numbering, feature titles, testimonial voice — but rewrite to the user's brief, never paste `Field Studio` or `Pretzel` verbatim into a different domain).
- Design tokens (colors, type stack, spacing scale) — lift as a palette starting point, then tune to the brief.
- Visual patterns (overlap, asymmetry, full-bleed blocks, grain overlay, numbered eyebrows).

You are the senior designer; the pattern is a sketch on the whiteboard, not the finished page.

## File index

| File | Good for |
|---|---|
| `landing-page.jsx` | Full marketing pages — 3 variants (editorial-serif / product-tech / minimal-startup). For just a hero, use `heroes.jsx`. |
| `heroes.jsx` | Hero blocks — multiple distinct hero patterns to lift. |
| `pricing.jsx` | 3-tier pricing tables with feature comparison. |
| `dashboard.jsx` | Admin / analytics dashboards with KPI strip + charts. |
| `chart-svg.jsx` | Pure-SVG chart patterns (line, bar, area). Use when cdnjs lib is unnecessary. |
| `data-table.jsx` | Dense tables with sorting/filter chips. |
| `chat-ui.jsx` | Chat interfaces, messaging threads. |
| `calendar.jsx` | Calendars, date pickers, availability grids. |
| `footers.jsx` | Site footers — multiple density levels. |
| `editorial-typography.jsx` | Long-form text, case studies, journal posts. |
| `glassmorphism.jsx` | Glass/blur surfaces — use sparingly, only when the tone calls for it. |
| `slide-deck.jsx` | Presentation slides — pitch decks, keynotes. |

## When NOT to use a pattern

- Brief explicitly overrides the pattern's tone (e.g. user wants brutalist, pattern is editorial-serif).
- The pattern's variants don't match the type. Fall back to `references/craft-directives.md` and build from first principles.
- Pure abstract / experimental briefs — patterns are baselines, not ceilings.

**Always run the anti-slop self-check after adapting.** Patterns can drag you toward the same aesthetic; anti-slop keeps you honest.
