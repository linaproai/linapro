module lina-plugins

go 1.25.0

require (
	lina-plugin-content-notice v0.0.0
	lina-plugin-demo-control v0.0.0
	lina-plugin-demo-source v0.0.0
	lina-plugin-monitor-loginlog v0.0.0
	lina-plugin-monitor-online v0.0.0
	lina-plugin-monitor-operlog v0.0.0
	lina-plugin-monitor-server v0.0.0
	lina-plugin-multi-tenant v0.0.0
	lina-plugin-org-center v0.0.0
)

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/emirpasic/gods/v2 v2.0.0-alpha // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogf/gf/contrib/drivers/pgsql/v2 v2.10.0 // indirect
	github.com/gogf/gf/contrib/drivers/sqlite/v2 v2.10.0 // indirect
	github.com/gogf/gf/v2 v2.10.1-0.20260422124457-cb7cfa58abb7 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grokify/html-strip-tags-go v0.1.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mssola/useragent v1.0.0 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.0.9 // indirect
	github.com/olekukonko/tablewriter v1.1.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/richardlehane/mscfb v1.0.6 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/shirou/gopsutil/v4 v4.26.3 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/excelize/v2 v2.10.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lina-core v0.0.0 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1 // indirect
)

replace lina-core => ../lina-core

replace lina-plugin-content-notice => ./content-notice

replace lina-plugin-demo-control => ./demo-control

replace lina-plugin-demo-source => ./plugin-demo-source

replace lina-plugin-monitor-loginlog => ./monitor-loginlog

replace lina-plugin-monitor-online => ./monitor-online

replace lina-plugin-monitor-operlog => ./monitor-operlog

replace lina-plugin-monitor-server => ./monitor-server

replace lina-plugin-multi-tenant => ./multi-tenant

replace lina-plugin-org-center => ./org-center
