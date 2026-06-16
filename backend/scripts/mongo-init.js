db = db.getSiblingDB("medical_agent");

db.createCollection("users");
db.createCollection("sessions");
db.createCollection("roles");
db.createCollection("knowledge_bases");
db.createCollection("documents");
db.createCollection("chunks");
db.createCollection("ingestion_jobs");
db.createCollection("conversations");
db.createCollection("messages");
db.createCollection("audit_logs");
db.createCollection("model_configs");

db.users.createIndex({ account: 1 }, { unique: true });
db.users.createIndex({ status: 1 });

db.sessions.createIndex({ tokenHash: 1 }, { unique: true });
db.sessions.createIndex({ userId: 1 });
db.sessions.createIndex({ expiresAt: 1 });

db.knowledge_bases.createIndex({ name: "text", description: "text" });
db.knowledge_bases.createIndex({ scenario: 1 });
db.knowledge_bases.createIndex({ tags: 1 });
db.knowledge_bases.createIndex({ department: 1 });
db.knowledge_bases.createIndex({ buildStatus: 1 });
db.knowledge_bases.createIndex({ updatedAt: -1 });

db.documents.createIndex({ knowledgeBaseId: 1 });
db.documents.createIndex({ status: 1 });
db.documents.createIndex({ createdAt: -1 });

db.chunks.createIndex({ knowledgeBaseId: 1 });
db.chunks.createIndex({ documentId: 1 });
db.chunks.createIndex({ vectorId: 1 }, { unique: true });

db.ingestion_jobs.createIndex({ status: 1 });
db.ingestion_jobs.createIndex({ documentId: 1 });
db.ingestion_jobs.createIndex({ updatedAt: -1 });

db.conversations.createIndex({ userId: 1, updatedAt: -1 });
db.conversations.createIndex({ title: "text" });

db.messages.createIndex({ conversationId: 1, createdAt: 1 });
db.messages.createIndex({ role: 1 });

db.audit_logs.createIndex({ actorId: 1, createdAt: -1 });
db.audit_logs.createIndex({ action: 1 });

db.model_configs.createIndex({ updatedAt: -1 });
