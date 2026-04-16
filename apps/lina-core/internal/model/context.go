package model

// Context is the business context for each request.
type Context struct {
	TokenId  string `json:"tokenId"`
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Status   int    `json:"status"`
}
