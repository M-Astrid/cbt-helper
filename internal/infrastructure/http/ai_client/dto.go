package ai_client

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	//MaxTokens int       `json:"max_tokens"`
}

type ResponseChoice struct {
	Message Message `json:"message"`
}

type ResponseBody struct {
	Choices []ResponseChoice `json:"choices"`
}
