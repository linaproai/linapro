// This file configures the project-wide GoFrame log handler used by the host.

package logger

import "github.com/gogf/gf/v2/os/glog"

// Configure applies the project-wide default GoFrame log handler.
func Configure(structured bool) {
	if structured {
		glog.SetDefaultHandler(glog.HandlerJson)
		return
	}
	glog.SetDefaultHandler(nil)
}
