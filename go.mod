module github.com/bonedaddy/mychurnero

go 1.14

require (
	github.com/jinzhu/gorm v1.9.16
	github.com/mattn/go-sqlite3 v1.14.2 // indirect
	github.com/monero-ecosystem/go-monero-rpc-client v0.0.0-20200124164006-0afb4abdfc3c
	github.com/segmentio/ksuid v1.0.3
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.2.0
	go.bobheadxi.dev/zapx/zapx v0.6.8
	go.uber.org/multierr v1.5.0
	go.uber.org/zap v1.10.0
	gopkg.in/yaml.v2 v2.2.2
	gorm.io/driver/sqlite v1.1.1
	gorm.io/gorm v1.20.1-0.20200904063544-f1216222284f
)

replace github.com/monero-ecosystem/go-monero-rpc-client v0.0.0-20200124164006-0afb4abdfc3c => github.com/bonedaddy/go-monero-rpc-client v0.0.0-20200904065722-23238e3895c4
