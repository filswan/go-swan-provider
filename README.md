# swan-miner

# build executable file for linux
GOOS=linux GOARCH=amd64 go build -v ./

# build executable file for mac
env GOOS=darwin GOARCH=amd64 go build

# before exution, please check configure file and ensure the configurations are all right

