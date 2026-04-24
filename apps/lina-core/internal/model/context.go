// This file defines the request-scoped business context shared across host
// middleware, controllers, and services.

package model

// Context is the business context for each request.
type Context struct {
	TokenId  string `json:"tokenId"`
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Status   int    `json:"status"`
	Locale   string `json:"locale"`
}
