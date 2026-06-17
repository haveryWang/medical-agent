# Agent Chat Specification

## Purpose
Provide an authenticated RAG chat workspace with conversation management, streaming answers, citations, answer details, and DeepSeek chat configuration boundaries.

## Requirements

### Requirement: Conversation workspace
The system SHALL provide an authenticated conversation workspace with conversation creation, search, history, active conversation display, and recycle bin entry points.

#### Scenario: Create new conversation
- **WHEN** the user clicks the new conversation action
- **THEN** the system creates a conversation and selects it as the active chat

#### Scenario: Search conversations
- **WHEN** the user enters text in the conversation search field
- **THEN** the system filters conversations by title and recent message content

### Requirement: Streaming answer generation
The system SHALL stream model-generated answer content to the frontend while the backend is receiving it from the model provider.

#### Scenario: Stream answer chunks
- **WHEN** the user sends a question in an active conversation
- **THEN** the frontend displays answer tokens or chunks progressively without waiting for the full answer

#### Scenario: Stream completes
- **WHEN** the model response finishes
- **THEN** the system persists the final assistant message, citations, timing, and retrieval metadata

#### Scenario: Stream fails
- **WHEN** generation fails after the stream has started
- **THEN** the frontend displays a recoverable failure state and the backend records the partial message status

### Requirement: RAG-grounded answers
The system SHALL use selected or eligible knowledge bases to retrieve relevant source chunks before calling the chat model.

#### Scenario: Answer includes sources
- **WHEN** retrieved chunks are used for an answer
- **THEN** the answer detail panel lists cited documents, relevance indicators, and source snippets

#### Scenario: No relevant context
- **WHEN** retrieval returns no chunks above the configured relevance threshold
- **THEN** the system tells the user that no reliable knowledge-base context was found and avoids fabricating citations

### Requirement: Answer detail panel
The system SHALL provide an answer detail panel for each assistant response.

#### Scenario: User opens answer details
- **WHEN** the user selects the answer detail action for an assistant message
- **THEN** the panel displays message id, response duration, knowledge references, retrieval strategy, and the prompt context sent to the model

### Requirement: DeepSeek chat configuration placeholder
The system SHALL read DeepSeek chat configuration from environment or deployment configuration and SHALL NOT hardcode the API key in source code.

#### Scenario: Missing DeepSeek API key
- **WHEN** the backend starts without a DeepSeek API key
- **THEN** startup documentation explains where to set it and DeepSeek chat requests fail with a clear configuration error

#### Scenario: DeepSeek chat model configured
- **WHEN** the operator provides a DeepSeek API key and chat model configuration
- **THEN** chat generation uses the configured DeepSeek service without exposing secrets to the frontend
