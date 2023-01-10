module swan-provider

go 1.16

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/Khan/genqlient v0.5.0
	github.com/fatih/color v1.13.0
	github.com/filecoin-project/go-jsonrpc v0.1.9 // indirect
	github.com/filswan/go-swan-lib v0.2.137
	github.com/gin-gonic/gin v1.7.4
	github.com/google/uuid v1.3.0
	github.com/itsjamie/gin-cors v0.0.0-20160420130702-97b4a9da7933
	github.com/joho/godotenv v1.4.0
	github.com/pkg/errors v0.9.1
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
