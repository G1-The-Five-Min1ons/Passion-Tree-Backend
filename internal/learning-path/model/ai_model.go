package model

type AIGeneratePathRequest struct {
	Topic string `json:"topic"`
}

type AIPathGenerationResponse struct {
	Topic       string   `json:"topic"`
	Result      string   `json:"result"`
	UsedContext []string `json:"used_context"`
}

type GeneratedPathResponse struct {
	Topic string          `json:"topic"`
	Nodes []GeneratedNode `json:"nodes"`
}

type GeneratedNode struct {
	Sequence int    `json:"sequence"`
	Title    string `json:"title"`
}