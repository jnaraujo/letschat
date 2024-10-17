package message

type AuthMessageClient struct {
	Username string `json:"username"`
	// Add stuff here like public key, ID, etc.
}

type AuthMessageServer struct {
	Status  string `json:"status"`
	Content string `json:"content"`
}
