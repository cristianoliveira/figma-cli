# Daylight

A calm, local-first todo app built with React, TypeScript, and Vite.

## Run

Requires Node.js 22.12+ and npm 11.

```sh
npm ci
npm run dev
```

Open **http://localhost:5173**. For a production bundle, run `npm run build`, then serve `dist/` on a static host. `npm run preview` previews that bundle locally.

## Features

- Create, edit, complete, reopen, and delete tasks.
- Projects, notes, due dates, and three priority levels.
- Today, Upcoming, Completed, project filters, and text search (⌘K / Ctrl+K).
- List or priority board layout; sort by due date, priority, or newest.
- Multiple file uploads and drag-and-drop: up to 10 files per task, 25 MB each.
- Image and animated GIF previews, native video playback, and file downloads.
- A Files & media collection linked to the original tasks.
- Browser persistence for tasks and binary attachments using IndexedDB.
- Responsive navigation, keyboard-accessible dialogs, reduced-motion support, and locally bundled fonts.

The first visit includes editable sample tasks. They are ordinary tasks: delete or replace them to make this workspace yours.

## Privacy and limits

Nothing is uploaded to a server. This is a **single-user, single-browser workspace**, not a cloud service. There is no login, cross-device sync, collaborative editing, or backup service. Use one tab at a time; concurrent tabs can overwrite each other's changes.

Clearing site data, changing browser/origin, or using private browsing affects persistence. Keep separate copies of important files. Storage limits vary by browser; failed saves retain the task draft and show an error.

GIF/image/video previews use revocable object URLs. Video formats and codecs depend on browser support. HTML, SVG, PDF, and other documents are download-only, not embedded or executed. Files are not malware-scanned: only open downloaded files you trust. Task deletion removes its attachment blobs from the workspace.

Fonts are bundled locally; the running app has no third-party requests. Dependency lifecycle scripts are disabled in `.npmrc`.

## Verify

```sh
npm test
npm run test:coverage
npm run build
npm run format:check
npm audit
```

Tests exercise task lifecycle, projects, filters, upload limits, persistence, failure recovery, and document downloads. `fake-indexeddb` substitutes the browser API in tests. Actual GIF/video decoding and reload persistence were also checked in Chromium.

## Source

- `src/App.tsx`: workspace composition and navigation.
- `src/TaskEditor.tsx`: task form, drop zone, and attachment editing.
- `src/AttachmentPreview.tsx`: safe media previews and downloads.
- `src/Modal.tsx`: native accessible dialog.
- `src/model.ts`: types, filtering, validation, dates, and starter tasks.
- `src/storage.ts`: transactional IndexedDB reads and writes.
- `src/styles.css`: responsive layout and component styles.
- `src/theme.css`: reference-inspired white/gray palette, blue actions, and semantic status colors.
