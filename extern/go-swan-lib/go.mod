module github.com/filswan/go-swan-lib

go 1.16

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/filecoin-project/boost v1.5.0
	github.com/filecoin-project/go-dagaggregator-unixfs v0.3.0
	github.com/filecoin-project/go-jsonrpc v0.1.9
	github.com/ipfs/go-blockservice v0.4.0
	github.com/ipfs/go-cid v0.2.0
	github.com/ipfs/go-ipfs-api v0.2.0
	github.com/ipfs/go-ipfs-blockstore v1.2.0
	github.com/ipfs/go-ipfs-exchange-offline v0.3.0
	github.com/ipfs/go-ipfs-files v0.1.1
	github.com/ipfs/go-merkledag v0.8.0
	github.com/mattn/go-isatty v0.0.16
	github.com/multiformats/go-multihash v0.2.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.8.1
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f
)

replace github.com/filecoin-project/filecoin-ffi => ../filecoin-ffi
