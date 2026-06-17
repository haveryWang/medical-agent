## Why

Operators need a lightweight place to capture fragmented review experience without polluting the existing RAG knowledge base, and they need a structured policy library that can present curated policy Excel data in a familiar category-first browsing interface. This change adds both product areas while preserving the current knowledge retrieval system as a separate capability.

## What Changes

- Add a "复盘笔记" section in the authenticated shell sidebar for collecting short operational notes.
- Persist review notes in an independent MongoDB collection or table that is physically separate from existing knowledge-base documents, chunks, and vector indexes.
- Record review-note creation time, exported state, and backend counts for unexported and total records.
- Alert or highlight in the UI when unexported review notes reach 15 records with the message "记录已达15条，可生成文档".
- Add a one-click generate-and-download action that exports the current export batch as a Markdown file and marks included records as exported.
- Add a "政策文件库" section that imports curated Excel policy data and displays it as policy cards or rows.
- Provide seven fixed policy categories: 国家医学中心、科技创新、医疗服务、医保医药、数智治理、改革监管、其他.
- Allow users to click a category tile or tab to filter the right-side policy list.
- Display each policy item with title, summary, date, and category tag.

## Capabilities

### New Capabilities

- `review-notes`: Capturing, counting, threshold prompting, Markdown exporting, and export-state tracking for physically isolated review notes.
- `policy-library`: Importing curated Excel policy records and browsing them by seven fixed policy categories.

### Modified Capabilities

- None. Existing accepted capabilities remain unchanged; review notes and policy library data SHALL stay separate from the current RAG knowledge-base retrieval system.

## Impact

- Adds backend APIs, store models, and MongoDB indexes for review notes and policy files.
- Adds frontend routes, sidebar navigation items, pages, filters, threshold prompts, and download flows.
- Adds Excel import support for policy library records, likely reusing existing spreadsheet parsing dependencies where appropriate without routing imported policies into RAG ingestion.
- Adds Markdown export generation for review notes.
- Adds tests for data isolation, note counts, threshold UI behavior, Markdown downloads, Excel import validation, category filtering, and authenticated access.
- Updates API, architecture, runbook, and task-log documentation.
