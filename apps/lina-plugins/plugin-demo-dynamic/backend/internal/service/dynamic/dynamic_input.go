// This file defines typed business inputs consumed by the dynamic plugin
// backend services.

package dynamicservice

// BackendSummaryInput defines the bridge context required by the backend
// summary business flow.
type BackendSummaryInput struct {
	PluginID      string
	PublicPath    string
	Access        string
	Permission    string
	Authenticated bool
	HasIdentity   bool
	Username      string
	IsSuperAdmin  bool
}

// DemoRecordListInput defines one paged demo-record query request.
type DemoRecordListInput struct {
	PageNum  int
	PageSize int
	Keyword  string
}

// DemoRecordMutationInput defines one demo-record create or update command.
type DemoRecordMutationInput struct {
	Title                   string `json:"title"`
	Content                 string `json:"content"`
	AttachmentName          string `json:"attachmentName"`
	AttachmentContentBase64 string `json:"attachmentContentBase64"`
	AttachmentContentType   string `json:"attachmentContentType"`
	RemoveAttachment        bool   `json:"removeAttachment"`
}

// HostCallDemoInput defines the request context required by the host-call
// demo business flow.
type HostCallDemoInput struct {
	PluginID    string
	Username    string
	RequestID   string
	RoutePath   string
	SkipNetwork bool
}
