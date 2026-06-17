# RAG Ingestion Retrieval Specification

## Purpose
Define document parsing, chunking, Qwen3 embedding generation, vector storage, retrieval, prompt assembly, and RAG observability behavior.

## Requirements

### Requirement: Document parsing
The system SHALL parse uploaded PDF, Word, Excel, Markdown, and text documents into normalized text sections for indexing.

#### Scenario: Parse supported document
- **WHEN** an ingestion job receives a supported document
- **THEN** the system extracts normalized text and document metadata suitable for chunking

#### Scenario: Parse failure
- **WHEN** text extraction fails
- **THEN** the ingestion job records a failed status and does not create partial vector entries as completed data

### Requirement: Chunking and metadata
The system SHALL split parsed text into searchable chunks with metadata linking each chunk to knowledge base, document, section, page or sheet when available, and original text offsets when available.

#### Scenario: Chunk document
- **WHEN** parsed text exceeds the configured chunk size
- **THEN** the system creates overlapping chunks using configured chunk size and overlap values

### Requirement: Qwen3 embedding generation
The system SHALL use Qwen3-Embedding as the embedding model for document chunks and query text.

#### Scenario: Qwen3 embedding configured
- **WHEN** Qwen3 embedding credentials, endpoint, and model configuration are present
- **THEN** ingestion jobs generate embeddings for chunks through Qwen3-Embedding and store vector ids with chunk metadata

#### Scenario: Qwen3 embedding unavailable
- **WHEN** Qwen3 embedding configuration is absent or the configured Qwen3-Embedding model is unavailable
- **THEN** ingestion jobs fail with a clear configuration error and do not silently skip vector indexing

### Requirement: Lightweight vector storage
The system SHALL use a lightweight vector database for similarity search and maintain MongoDB as the source of truth for non-vector data.

#### Scenario: Store vectors
- **WHEN** a chunk embedding is generated
- **THEN** the system stores the vector in the vector database and stores the vector id plus metadata in MongoDB

#### Scenario: Delete or disable document
- **WHEN** a document is deleted or disabled
- **THEN** the system removes or excludes related vectors from retrieval and preserves required audit metadata in MongoDB

### Requirement: Retrieval strategy
The system SHALL retrieve candidate chunks using vector similarity and configurable filters, then return a ranked context set for prompt assembly.

#### Scenario: Retrieve with filters
- **WHEN** a chat request includes selected knowledge bases or department scope
- **THEN** retrieval only considers chunks permitted by those filters

#### Scenario: Rank candidates
- **WHEN** multiple chunks match a query
- **THEN** the system returns the top configured number of chunks with scores and metadata

### Requirement: Prompt assembly
The system SHALL assemble model prompts from user question, conversation context, retrieved chunks, and system instructions that require grounded answers.

#### Scenario: Build grounded prompt
- **WHEN** relevant chunks are found
- **THEN** the prompt includes source identifiers and tells the model to answer using the provided context and cite sources

### Requirement: Observability metrics
The system SHALL record retrieval and generation metadata for troubleshooting and answer detail display.

#### Scenario: Record RAG metadata
- **WHEN** an answer is generated
- **THEN** the system stores retrieval latency, selected chunk ids, vector scores, model name, token usage when available, and total response duration
