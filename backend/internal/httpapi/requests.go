package httpapi

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type createConversationRequest struct {
	Title            string   `json:"title"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
}

type updateConversationRequest struct {
	Title            *string  `json:"title"`
	Status           *string  `json:"status"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
}

type streamMessageRequest struct {
	Content          string   `json:"content"`
	KnowledgeBaseIDs []string `json:"knowledgeBaseIds"`
}

type saveModelConfigRequest struct {
	DeepSeekBaseURL        string `json:"deepSeekBaseUrl"`
	DeepSeekAPIKey         string `json:"deepSeekAPIKey"`
	DeepSeekChatModel      string `json:"deepSeekChatModel"`
	QwenEmbeddingBaseURL   string `json:"qwenEmbeddingBaseUrl"`
	QwenEmbeddingAPIKey    string `json:"qwenEmbeddingAPIKey"`
	QwenEmbeddingModel     string `json:"qwenEmbeddingModel"`
	QwenEmbeddingDimension int    `json:"qwenEmbeddingDimension"`
}
