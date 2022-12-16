module swan-provider

go 1.16

require (
	github.com/BurntSushi/toml v1.1.0
	github.com/Khan/genqlient v0.5.0 // indirect
	github.com/filecoin-project/go-jsonrpc v0.1.9 // indirect
	github.com/filswan/go-swan-lib v0.2.131
	github.com/gin-gonic/gin v1.7.4
	github.com/graphql-go/graphql v0.8.0 // indirect
	github.com/itsjamie/gin-cors v0.0.0-20160420130702-97b4a9da7933
	github.com/joho/godotenv v1.4.0
)

replace github.com/filswan/go-swan-lib => ./extern/go-swan-lib
