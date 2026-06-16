package httpapi

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type createConversationRequest struct {
	Title            string   `json:"title"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
}

type streamMessageRequest struct {
	Content          string   `json:"content"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
}
