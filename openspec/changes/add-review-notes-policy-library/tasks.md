## 1. Backend Data Model And Permissions

- [x] 1.1 Add review-note, review-note export, policy document, and policy import batch models in `backend/internal/models`.
- [x] 1.2 Add MongoDB indexes for review-note actor/export state/created time and policy category/date/title fields.
- [x] 1.3 Add store methods for creating, listing, counting, exporting, and batch-updating review notes in a dedicated store file.
- [x] 1.4 Add store methods for importing, listing, and category-filtering policy documents in a dedicated store file.
- [x] 1.5 Seed or migrate administrator permissions for review-note write/export and policy import actions without reusing knowledge-base permissions.

## 2. Review Notes Backend API

- [x] 2.1 Add request/response structs for review-note creation, listing, counts, and export results.
- [x] 2.2 Implement authenticated endpoints for creating review notes, listing recent notes, and retrieving total/unexported counts.
- [x] 2.3 Implement Markdown export endpoint that claims unexported notes, writes export batch metadata, marks included notes as exported, and returns an attachment response.
- [x] 2.4 Validate empty review-note submissions and empty export requests with the shared API error format.
- [x] 2.5 Add backend tests for note creation timestamps, counts, export state transitions, Markdown content, and physical isolation from knowledge-base collections.

## 3. Policy Library Backend API

- [x] 3.1 Add fixed category constants and validation for 国家医学中心、科技创新、医疗服务、医保医药、数智治理、改革监管、其他.
- [x] 3.2 Implement policy Excel import parsing for title, summary, date, and category columns with documented column aliases.
- [x] 3.3 Implement authenticated policy list endpoint with category filtering and empty-result behavior.
- [x] 3.4 Implement permitted policy import endpoint that records import batch counts and validation errors.
- [x] 3.5 Add backend tests for valid import, unsupported category rows, invalid files, category filtering, display fields, and physical isolation from RAG ingestion/vector storage.

## 4. Frontend Navigation And API Client

- [x] 4.1 Extend the authenticated shell navigation with "复盘笔记" and "政策文件库" entries while preserving existing chat and knowledge-base routes.
- [x] 4.2 Add protected routes for `/review-notes` and `/policies`.
- [x] 4.3 Add API client methods for review-note create/list/count/export and policy list/import.
- [x] 4.4 Add frontend download handling for Markdown export with the backend-provided filename.

## 5. Review Notes Frontend

- [x] 5.1 Create review-note feature modules for note submission, recent note list, count loading, and export action state.
- [x] 5.2 Build the "复盘笔记" page with a note input form, submit action, total/unexported counts, and recent note list showing created time and exported state.
- [x] 5.3 Display an automatic modal, banner, or high-visibility prompt with "记录已达15条，可生成文档" when unexported count is at least 15.
- [x] 5.4 Wire "一键生成并下载" to the Markdown export endpoint and refresh counts/list after export.
- [x] 5.5 Add frontend tests for validation, threshold prompt visibility, export action, and count refresh.

## 6. Policy Library Frontend

- [x] 6.1 Create policy-library feature modules for fixed categories, category selection, policy list loading, and optional import state.
- [x] 6.2 Build the "政策文件库" page with a left fixed category area and right policy list matching the reference page structure.
- [x] 6.3 Render each policy record with title, summary, date, and category tag.
- [x] 6.4 Add category click filtering and an empty state for categories with no records.
- [x] 6.5 Add permitted Excel import UI if the authenticated user has policy import permission.
- [x] 6.6 Add frontend tests for fixed category rendering, category filtering, empty state, policy fields, and import validation feedback.

## 7. Documentation And Verification

- [x] 7.1 Update API documentation with review-note and policy-library endpoints, request/response shapes, and download behavior.
- [x] 7.2 Update architecture documentation to describe physical isolation from knowledge-base and RAG storage.
- [x] 7.3 Update runbooks with policy Excel template/column aliases and review-note Markdown export behavior.
- [x] 7.4 Append implementation evidence to `docs/records/task-log.md`.
- [x] 7.5 Run backend tests with `CGO_ENABLED=0 go test ./...` and record results.
- [x] 7.6 Run frontend checks/build from `frontend/` and record results.
- [x] 7.7 Manually verify navigation, note submission, 15-note prompt, Markdown download, policy import, and policy category filtering in the browser.
