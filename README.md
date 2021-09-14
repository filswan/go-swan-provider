# swan-miner

# build executable file for linux
GOOS=linux GOARCH=amd64 go build -v ./

# build executable file for mac
env GOOS=darwin GOARCH=amd64 go build -v ./

# put the bin file to destination

# create config folder
in the same directory as the bin file

# put config.toml under config folder
the source file is:
go-swan-provider/config/config.toml

# edit config.toml with right values

