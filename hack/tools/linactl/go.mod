module linactl

go 1.25.0

require (
	github.com/lib/pq v1.10.9
	gopkg.in/yaml.v3 v3.0.1
	lina-core v0.0.0
)

require (
	github.com/gogf/gf/v2 v2.10.1 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace lina-core => ../../../apps/lina-core
