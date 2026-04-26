package v1

import "github.com/gogf/gf/v2/frame/g"

// SummaryReq is the request for querying plugin-demo-source summary.
type SummaryReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/summary" method:"get" tags:"Source Plugin Demo" summary:"Query source plugin example summary" dc:"Return summary copy for the plugin-demo-source page to verify that a source plugin menu page can read backend API data." permission:"plugin-demo-source:example:view"`
}

// SummaryRes is the response for querying plugin-demo-source summary.
type SummaryRes struct {
	Message string `json:"message" dc:"A brief introduction copy used for page display, from the plugin backend interface" eg:"This is a brief introduction from the plugin-demo-source interface, which is used to verify that the source plugin menu page can read the plugin backend data."`
}
