## ADDED Requirements

### Requirement: Knowledge-base list
The system SHALL provide a knowledge-base list with filters for scenario, tag, department, name or description, build status, update time, and row actions.

#### Scenario: Filter knowledge bases
- **WHEN** the user applies scenario, tag, department, or keyword filters
- **THEN** the system returns matching knowledge bases with pagination metadata

#### Scenario: Display build status
- **WHEN** the list renders each knowledge base
- **THEN** it displays document count, build status, last update time, and available actions

### Requirement: Knowledge-base metadata management
The system SHALL allow permitted users to create, edit, and disable knowledge bases with scenario, tag, department, description, and retrieval configuration metadata.

#### Scenario: Create knowledge base
- **WHEN** a permitted user submits valid knowledge-base metadata
- **THEN** the system persists the knowledge base in MongoDB and makes it available for document upload

#### Scenario: Disable knowledge base
- **WHEN** a permitted user disables a knowledge base
- **THEN** the system excludes it from retrieval while preserving documents and audit history

### Requirement: Document upload
The system SHALL allow permitted users to upload supported document files to a knowledge base.

#### Scenario: Upload supported file
- **WHEN** the user uploads a PDF, Word, Excel, Markdown, or text file under the configured size limit
- **THEN** the system stores document metadata, stores the raw file or configured object reference, and creates an ingestion job

#### Scenario: Upload unsupported file
- **WHEN** the user uploads an unsupported file type or oversized file
- **THEN** the system rejects the upload with a validation error before ingestion begins

### Requirement: Ingestion status tracking
The system SHALL expose document ingestion status from upload through parsing, chunking, embedding, indexing, completion, or failure.

#### Scenario: Ingestion succeeds
- **WHEN** a document is parsed, chunked, embedded, and indexed successfully
- **THEN** the document and knowledge base show a completed build status

#### Scenario: Ingestion fails
- **WHEN** parsing, embedding, or vector indexing fails
- **THEN** the document shows a failed status with a reason and allows a permitted user to retry

### Requirement: Source governance
The system SHALL record audit events for knowledge-base creation, update, upload, deletion, indexing, retry, and permission-sensitive operations.

#### Scenario: Audit management operation
- **WHEN** a permitted user changes knowledge-base metadata or uploads a document
- **THEN** the system stores an audit event with actor, action, target, timestamp, and result
