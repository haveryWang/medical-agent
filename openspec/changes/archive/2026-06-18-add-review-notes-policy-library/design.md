## Context

The current system is a React + Ant Design frontend and Go + chi backend with MongoDB as canonical storage and Qdrant reserved for RAG vectors. Accepted capabilities cover authentication, chat, knowledge-base management, RAG ingestion/retrieval, and documentation. This change adds two authenticated product areas:

- "复盘笔记": a fragmented experience collector that must remain physically isolated from existing knowledge-base documents, chunks, ingestion jobs, and vector retrieval.
- "政策文件库": a curated policy browsing library loaded from a structured Excel file and displayed using seven fixed policy categories, inspired by the provided reference image.

The existing Go backend already includes Excel parsing dependencies for ingestion. The implementation can reuse parsing libraries, but policy Excel import must be a separate domain flow and must not create knowledge-base documents, chunks, ingestion jobs, or Qdrant vectors.

## Goals / Non-Goals

**Goals:**

- Add authenticated frontend navigation for "复盘笔记" and "政策文件库", with "复盘笔记" available from the left-side application navigation requested by the user.
- Store review notes in dedicated MongoDB collections with creation timestamps, export status, export batch metadata, and count APIs.
- Prompt users when unexported review-note count reaches 15 records, then allow one-click Markdown generation and direct browser download.
- Use unexported records as the export batch by default. The prompt threshold is 15; if more than 15 unexported records exist, all currently selected unexported records are included unless the product later adds a configurable batch size.
- Add policy Excel import into dedicated policy collections and display title, summary, date, and category tag.
- Enforce exactly seven policy categories: 国家医学中心、科技创新、医疗服务、医保医药、数智治理、改革监管、其他.
- Keep the current RAG knowledge-base retrieval behavior unchanged.

**Non-Goals:**

- Do not include review notes or policy library records in vector retrieval, chat citations, or knowledge-base document lists.
- Do not build a full policy search engine with relevance ranking, department taxonomy trees, or web crawling.
- Do not edit generated Markdown before download in this change.
- Do not implement collaborative review workflows, approvals, or rich-text note editing.
- Do not require a new external storage system.

## Decisions

### Decision: Dedicated MongoDB collections for new domains

Create dedicated collections such as:

- `review_notes`: content, optional title/tags/source context, actor id, exported flag, export batch id, created/updated/exported timestamps.
- `review_note_exports`: export id, actor id, included note ids, note count, Markdown filename, created timestamp.
- `policy_documents`: title, summary, publish date, category, source/issuer fields if present, import batch id, row checksum, created/updated timestamps.
- `policy_import_batches`: filename, actor id, imported count, skipped count, validation errors, created timestamp.

Rationale: separate collections make the physical isolation requirement auditable and avoid accidental coupling with `documents`, `chunks`, `ingestion_jobs`, and Qdrant payloads.

Alternatives considered:

- Store review notes as knowledge-base documents: rejected because it violates the physical isolation requirement and would expose notes to RAG retrieval.
- Store policies in the existing knowledge-base upload pipeline: rejected because the desired UI is curated browsing, not semantic retrieval.

### Decision: Review-note threshold uses unexported count

The backend returns both total and unexported counts, but the threshold prompt is based on unexported count. Export marks included notes as exported and associates them with an export batch.

Rationale: unexported count prevents users from repeatedly seeing the same threshold prompt after they generate a Markdown file. The request allowed "未导出(或总计)"; unexported count better matches the export workflow.

Alternatives considered:

- Use total count: simpler, but the prompt would remain true forever once the system has 15 lifetime records.
- Hard-code 30 exported records: the request contains both "15条" and "这30条"; this design resolves the conflict by treating 15 as the prompt threshold and exporting the current unexported batch.

### Decision: Markdown export is generated on demand by the backend

Add an authenticated endpoint that atomically claims the current unexported review-note batch, renders Markdown with a title, generated time, count, and chronological note sections, then returns it with `Content-Type: text/markdown` and `Content-Disposition: attachment`.

Rationale: backend generation avoids frontend/client clock inconsistencies and ensures export status is updated consistently with the downloaded content.

Alternatives considered:

- Generate Markdown entirely in the browser: easier UI work, but weaker auditability and no reliable export-state transition.
- Pre-generate files in storage: unnecessary until export volumes or retention requirements grow.

### Decision: Policy library import is structured and category validated

Add an authenticated Excel import endpoint that accepts a curated spreadsheet with at least title, summary, date, and category columns. The importer normalizes known Chinese column names, validates category values against the seven-category allowlist, rejects rows without required display fields, and records row-level validation counts.

Rationale: the user already has curated Excel data and wants display parity with the reference page. Strict category validation keeps the left filter stable and avoids dropdown/taxonomy drift.

Alternatives considered:

- Free-form category values from Excel: flexible, but violates the fixed seven-tag requirement.
- Seed policy documents by editing code: fast for a demo, but poor for future data updates.

### Decision: Frontend keeps list interactions simple

Add pages under `frontend/src/pages/ReviewNotesPage.jsx` and `frontend/src/pages/PolicyLibraryPage.jsx`, feature modules under `frontend/src/features/reviewNotes/` and `frontend/src/features/policyLibrary/`, and API methods in `frontend/src/api/client.js`.

Review notes UI includes:

- Note entry form with submit action.
- Recent note list with created time and exported state.
- Count display for total and unexported notes.
- Automatic modal, banner, or high-visibility prompt when unexported count is at least 15.
- "一键生成并下载" action wired to backend Markdown download.

Policy library UI includes:

- Left fixed category tiles or buttons for the seven categories, laid out without a dropdown.
- Right list of policy records with title, summary, date, and category tag.
- Optional keyword and import controls if permitted by user permissions.

Rationale: these pages fit the existing React/Ant Design approach while matching the requested policy library visual model.

Alternatives considered:

- Reuse knowledge-base table components directly: rejected because the policy library is browse-first and category-first, while the knowledge-base table is source-management-first.

### Decision: Permissions align with existing authenticated shell

All authenticated users can view policy records and create review notes unless existing role rules require stricter access. Importing policy Excel and exporting review notes should use explicit write permissions, for example `policy:write` and `review_notes:write`, seeded for administrator users.

Rationale: the existing backend already has permission middleware. Separate permissions make future governance easier without changing the user-facing requirements.

Alternatives considered:

- Use `knowledge:write` for policy import and review-note export: rejected because it blurs domain boundaries with the knowledge base.

## Risks / Trade-offs

- [Risk] The phrase "这30条记录" conflicts with the 15-record threshold. → Mitigation: document and implement 15 as the prompt threshold and export all currently unexported notes; keep batch sizing configurable later if product confirms a hard 30-record export requirement.
- [Risk] Policy Excel column names may vary. → Mitigation: support a small set of documented aliases for title, summary, date, and category; return row-level validation errors for unknown or incomplete rows.
- [Risk] Users may expect policies to appear in chat retrieval. → Mitigation: label the policy library as a browsing/import display feature and do not write policy rows to RAG collections or Qdrant.
- [Risk] Export status could be marked before a failed browser download. → Mitigation: generate and mark the export in one backend request; provide export history or retry by export batch if implementation time allows.
- [Risk] Adding a left-side navigation changes the current top-nav layout. → Mitigation: preserve existing routes and labels while adding the requested left navigation affordance in the authenticated shell.

## Migration Plan

1. Add MongoDB models, repositories, and indexes for review notes, review-note exports, policy documents, and policy import batches.
2. Seed or migrate administrator permissions for review-note and policy-library write actions.
3. Add backend endpoints under `/api/v1/review-notes` and `/api/v1/policies`.
4. Add frontend API methods, routes, navigation items, and pages.
5. Add documentation for APIs, Excel template requirements, Markdown export behavior, and physical isolation from RAG.
6. Deploy with existing MongoDB; no vector database migration is required.

Rollback removes the new routes/pages and leaves existing RAG collections untouched. New collections can be retained for audit or dropped explicitly after confirming no export/import data is needed.

## Open Questions

- Should Markdown export include exactly the first 30 unexported notes once the count reaches 15, or should it include all unexported notes as designed here?
- Which user roles beyond administrator should be allowed to import policy Excel files and export review-note Markdown files?
- What are the exact Excel column headers in the curated policy file?
