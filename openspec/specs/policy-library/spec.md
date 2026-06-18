# Policy Library Specification

## Purpose
Define authenticated policy-library navigation, fixed category browsing, policy Excel import, policy list filtering, and physical isolation from knowledge-base RAG storage.

## Requirements

### Requirement: Policy library navigation
The system SHALL provide an authenticated "政策文件库" navigation entry from the application shell.

#### Scenario: Open policy library page
- **WHEN** an authenticated user clicks the "政策文件库" navigation entry
- **THEN** the system displays the policy library page without navigating to the knowledge-base management page

### Requirement: Fixed policy categories
The policy library SHALL use exactly eight fixed categories: 国家医学中心、科技创新、医疗服务、医保医药、数智治理、改革监管、国际合作、其他.

#### Scenario: Render fixed categories
- **WHEN** the policy library page loads
- **THEN** the left category area displays all eight fixed categories as directly clickable options without a dropdown

#### Scenario: Reject unsupported category
- **WHEN** imported policy data contains a category outside the eight fixed categories
- **THEN** the system rejects or skips that row with a validation error and does not store it as a valid policy record

### Requirement: Policy Excel import
The system SHALL allow permitted users to import curated policy records from Excel into dedicated policy-library storage.

#### Scenario: Import valid Excel rows
- **WHEN** a permitted user imports an Excel file with valid title, summary, date, and category fields
- **THEN** the system stores policy records in policy-library storage and records the import batch result

#### Scenario: Normalize import dates
- **WHEN** imported policy data contains Excel date cells or loose date text such as `2026/6/6`, `2026.6.6`, `20260606`, `2026年6月6日`, or `2026/6`
- **THEN** the system normalizes dates to `YYYY-MM-DD` or `YYYY-MM` before storing records so month aggregation can include them

#### Scenario: Reject invalid Excel file
- **WHEN** a permitted user imports a non-Excel file or an Excel file without required display fields
- **THEN** the system rejects the import with a validation error and does not store incomplete policy records

### Requirement: Policy library physical isolation
The system SHALL keep policy library records separate from existing knowledge-base ingestion and vector retrieval storage.

#### Scenario: Store policies separately from knowledge base
- **WHEN** policy records are imported
- **THEN** the system stores them in policy-library storage and does not create knowledge-base documents, chunks, ingestion jobs, or vector records

#### Scenario: Exclude policies from RAG retrieval
- **WHEN** a chat request performs RAG retrieval
- **THEN** the system does not retrieve policy-library records as citations or prompt context unless a future accepted capability explicitly adds that behavior

### Requirement: Policy list display
The system SHALL display policy records with title, summary, date, category tag, and optional interpretation text.

#### Scenario: Display imported policy record
- **WHEN** the policy library list renders an imported policy record
- **THEN** the list item displays the record title, summary, date, and category tag

#### Scenario: Expand long policy text
- **WHEN** a policy summary or interpretation is longer than the default preview area
- **THEN** the UI provides an action to expand and show the full text

### Requirement: Policy list filtering
The system SHALL filter the policy list by selected fixed category, month, and policy title keyword.

#### Scenario: Filter by category
- **WHEN** an authenticated user clicks a fixed policy category
- **THEN** the right-side policy list displays records from that category and hides records from other categories

#### Scenario: Filter by month
- **WHEN** an authenticated user selects a month from the date aggregation filter
- **THEN** the right-side policy list displays records whose normalized date starts with that `YYYY-MM` month

#### Scenario: Search by policy title
- **WHEN** an authenticated user enters a name keyword in the policy-library filter bar
- **THEN** the system returns policy records whose title fuzzy-matches the keyword

#### Scenario: Debounce keyword search
- **WHEN** an authenticated user is typing a policy title keyword
- **THEN** the system delays automatic list reload briefly so it does not call the list API for every keystroke

#### Scenario: Show empty category state
- **WHEN** an authenticated user selects a category with no policy records
- **THEN** the system displays an empty state for the right-side policy list
