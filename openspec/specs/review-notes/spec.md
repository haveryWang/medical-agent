# Review Notes Specification

## Purpose
Define authenticated fragmented review-note capture, threshold prompting, Markdown export, and physical isolation from knowledge-base RAG storage.

## Requirements

### Requirement: Review notes navigation
The system SHALL provide an authenticated "复盘笔记" navigation entry from the application shell.

#### Scenario: Open review notes page
- **WHEN** an authenticated user clicks the "复盘笔记" navigation entry
- **THEN** the system displays the review notes collection page without navigating to the knowledge-base management page

### Requirement: Review note capture
The system SHALL allow authenticated users to submit fragmented review-note content and persist it with creation time.

#### Scenario: Submit review note
- **WHEN** an authenticated user submits non-empty review-note content
- **THEN** the system stores the note with content, actor, exported state set to false, and created time

#### Scenario: Reject empty review note
- **WHEN** an authenticated user submits empty or whitespace-only review-note content
- **THEN** the system rejects the submission with a validation error and does not create a record

### Requirement: Review note physical isolation
The system SHALL store review notes outside existing knowledge-base document, chunk, ingestion, and vector retrieval storage.

#### Scenario: Store note separately from knowledge base
- **WHEN** a review note is created
- **THEN** the system stores it in review-note storage and does not create a knowledge-base document, chunk, ingestion job, or vector record

#### Scenario: Exclude notes from RAG retrieval
- **WHEN** a chat request performs RAG retrieval
- **THEN** the system does not retrieve review-note content as a citation or prompt context

### Requirement: Review note counts
The system SHALL expose review-note counts for total records and currently unexported records.

#### Scenario: Count notes after submission
- **WHEN** an authenticated user creates a review note
- **THEN** the backend count response includes the updated total count and unexported count

#### Scenario: Count notes after export
- **WHEN** review notes are exported successfully
- **THEN** the backend count response reduces the unexported count by the number of exported notes while preserving total count

### Requirement: Review note threshold prompt
The system SHALL alert or highlight that a document can be generated when the unexported review-note count reaches 15.

#### Scenario: Threshold reached
- **WHEN** the unexported review-note count is at least 15
- **THEN** the review notes UI displays "记录已达15条，可生成文档" and provides a "一键生成并下载" action

#### Scenario: Threshold not reached
- **WHEN** the unexported review-note count is less than 15
- **THEN** the review notes UI does not show the threshold-ready prompt

### Requirement: Review note Markdown export
The system SHALL generate a downloadable Markdown file from the current unexported review-note batch.

#### Scenario: Generate and download Markdown
- **WHEN** an authenticated user clicks "一键生成并下载" while unexported notes are available
- **THEN** the system returns a Markdown file download containing the exported notes, their creation times, and a generated timestamp

#### Scenario: Mark notes as exported
- **WHEN** the Markdown export is generated
- **THEN** the system marks the included notes as exported and records an export batch

#### Scenario: Prevent empty export
- **WHEN** an authenticated user requests Markdown export with no unexported review notes
- **THEN** the system rejects the request with a validation error and does not create an export batch
