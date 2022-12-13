package ipfs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/filecoin-project/go-dagaggregator-unixfs"
	"github.com/filecoin-project/go-dagaggregator-unixfs/lib/rambs"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	ipfsapi "github.com/ipfs/go-ipfs-api"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	exchangeoffline "github.com/ipfs/go-ipfs-exchange-offline"
	ipfsfiles "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-merkledag"
	"github.com/mattn/go-isatty"
	"github.com/multiformats/go-multihash"
	"golang.org/x/sys/unix"
	"golang.org/x/xerrors"
)

type opts struct {
	IpfsAPI            string `getopt:"--ipfs-api             A read/write IPFS API URL"`
	IpfsAPIMaxWorkers  uint   `getopt:"--ipfs-api-max-workers Max amount of parallel API requests"`
	IpfsAPITimeoutSecs uint   `getopt:"--ipfs-api-timeout     Max amount of seconds for a single API operation"`
	ShowProgress       bool   `getopt:"--show-progress        Print progress on STDERR, default when a TTY"`
}

func MergeFiles2CarFile(apiUrl string, cidStrs []string) (*string, error) {
	opts := &opts{
		ShowProgress:       isatty.IsTerminal(os.Stderr.Fd()),
		IpfsAPIMaxWorkers:  8,
		IpfsAPITimeoutSecs: 300,
		IpfsAPI:            apiUrl,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, unix.SIGINT, unix.SIGTERM, unix.SIGHUP)
		<-sigs
		cancel()
	}()

	if len(cidStrs) == 0 {
		s := bufio.NewScanner(os.Stdin)
		s.Split(bufio.ScanWords)
		for s.Scan() {
			cidStrs = append(cidStrs, s.Text())
		}
	}

	cset := cid.NewSet()
	toAgg := make([]dagaggregator.AggregateDagEntry, 0, len(cidStrs))
	for _, cs := range cidStrs {
		c, err := cid.Parse(cs)
		if err != nil {
			logs.GetLogger().Error("unable to parse '%s': %s", cs, err)
			return nil, err
		}
		if cset.Visit(c) {
			toAgg = append(toAgg, dagaggregator.AggregateDagEntry{RootCid: c})
		}
	}

	ramBs := new(rambs.RamBs)
	ramDs := merkledag.NewDAGService(blockservice.New(ramBs, exchangeoffline.Exchange(ramBs)))
	root, entries, err := dagaggregator.Aggregate(ctx, ramDs, toAgg)
	if err != nil {
		logs.GetLogger().Error("aggregation failed: %s", err)
		return nil, err
	}

	if err := writeoutBlocks(ctx, opts, ramBs); err != nil {
		logs.GetLogger().Error("writing newly created dag to IPFS API failed: %s", err)
		return nil, err
	}

	akc, _ := ramBs.AllKeysChan(ctx)
	dataCid := root.String()
	logs.GetLogger().Info("aggregation finished, aggregateRoot: ", root, ", data cid:", dataCid, ", totalManifestEntries: ", len(entries), ", newIntermediateBlocks: ", len(akc))
	return &dataCid, nil
}

func statSources(externalCtx context.Context, opts *opts, toAgg []dagaggregator.AggregateDagEntry) error {

	type dagStat struct {
		Size      uint64
		NumBlocks uint64
	}

	innerCtx, shutdownWorkers := context.WithCancel(externalCtx)
	defer shutdownWorkers()

	dagsDone := new(uint64)

	// channel of toAgg indexes to work on
	workCh := make(chan int, len(toAgg))
	for i := range toAgg {
		workCh <- i
	}
	close(workCh)

	finishCh := make(chan struct{}, 1)
	maxWorkers := opts.IpfsAPIMaxWorkers
	errCh := make(chan error, maxWorkers)

	var wg sync.WaitGroup

	for maxWorkers > 0 {
		maxWorkers--
		wg.Add(1)
		go func() {
			defer wg.Done()
			api := ipfsapi.NewShell(opts.IpfsAPI)
			api.SetTimeout(time.Second * time.Duration(opts.IpfsAPITimeoutSecs))

			for {
				toAggIdx, chanOpen := <-workCh
				if !chanOpen {
					select {
					case finishCh <- struct{}{}:
					default:
						// if we can't signal feeder is done - someone else already did
					}
					return
				}

				ds := new(dagStat)
				err := api.Request("dag/stat").Arguments(toAgg[toAggIdx].RootCid.String()).Option("progress", "false").Exec(innerCtx, ds)
				if err != nil {
					errCh <- err
					return
				}

				toAgg[toAggIdx].UniqueBlockCount = ds.NumBlocks
				toAgg[toAggIdx].UniqueBlockCumulativeSize = ds.Size

				if opts.ShowProgress {
					atomic.AddUint64(dagsDone, 1)
				}
			}
		}()
	}

	var lastPct uint64
	dagsTotal := uint64(len(toAgg))
	var progressTick <-chan time.Time
	if opts.ShowProgress {
		fmt.Fprint(os.Stderr, "0% of dags analyzed\r")
		t := time.NewTicker(250 * time.Millisecond)
		progressTick = t.C
		defer t.Stop()
	}

	var workerError error
watchdog:
	for {
		select {

		case <-finishCh:
			break watchdog

		case <-externalCtx.Done():
			break watchdog

		case workerError = <-errCh:
			shutdownWorkers()
			break watchdog

		case <-progressTick:
			curPct := 100 * atomic.LoadUint64(dagsDone) / dagsTotal
			if curPct != lastPct {
				lastPct = curPct
				fmt.Fprintf(os.Stderr, "%d%% of dags analyzed\r", lastPct)
			}
		}
	}

	wg.Wait()
	close(errCh) // closing a buffered channel keeps any buffered values for <-

	if workerError != nil {
		return workerError
	}
	if err := <-errCh; err != nil {
		return err
	}
	return externalCtx.Err()
}

// pulls cids from an AllKeysChan and sends them concurrently via multiple workers to an API
func writeoutBlocks(externalCtx context.Context, opts *opts, bs blockstore.Blockstore) error {

	innerCtx, shutdownWorkers := context.WithCancel(externalCtx)
	defer shutdownWorkers()

	akc, err := bs.AllKeysChan(innerCtx)
	if err != nil {
		return err
	}

	maxWorkers := opts.IpfsAPIMaxWorkers
	finishCh := make(chan struct{}, 1)
	errCh := make(chan error, maxWorkers)

	// WaitGroup as we want everyone to fully "quit" before we return
	var wg sync.WaitGroup
	blocksDone := new(uint64)

	for i := uint(0); i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			api := ipfsapi.NewShell(opts.IpfsAPI)
			api.SetTimeout(time.Second * time.Duration(opts.IpfsAPITimeoutSecs))

			for {
				select {

				case <-innerCtx.Done():
					// something caused us to stop, whatever it is parent knows why
					return

				case c, chanOpen := <-akc:

					if !chanOpen {
						select {
						case finishCh <- struct{}{}:
						default:
							// if we can't signal feeder is done - someone else already did
						}
						return
					}

					blk, err := bs.Get(c)
					if err != nil {
						errCh <- err
						return
					}

					// copied entirety of ipfsapi.BlockPut() to be able to pass in our own ctx 🤮
					res := new(struct{ Key string })
					err = api.Request("block/put").
						Option("format", cid.CodecToStr[c.Prefix().Codec]).
						Option("mhtype", multihash.Codes[c.Prefix().MhType]).
						Option("mhlen", c.Prefix().MhLength).
						Body(
							ipfsfiles.NewMultiFileReader(
								ipfsfiles.NewSliceDirectory([]ipfsfiles.DirEntry{
									ipfsfiles.FileEntry(
										"",
										ipfsfiles.NewBytesFile(blk.RawData()),
									),
								}),
								true,
							),
						).
						Exec(innerCtx, res)
					// end of 🤮

					if err != nil {
						errCh <- err
						return
					}

					if res.Key != c.String() {
						errCh <- xerrors.Errorf("unexpected cid mismatch after /block/put: expected %s but got %s", c, res.Key)
						return
					}

					if opts.ShowProgress {
						atomic.AddUint64(blocksDone, 1)
					}
				}
			}
		}()
	}

	var blocksTotal, lastPct uint64
	var progressTick <-chan time.Time
	if opts.ShowProgress {
		// this works because of how AllKeysChan behaves on rambs
		blocksTotal = uint64(len(akc))
		fmt.Fprint(os.Stderr, "0% of blocks written\r")
		t := time.NewTicker(250 * time.Millisecond)
		progressTick = t.C
		defer t.Stop()
	}

	var workerError error
watchdog:
	for {
		select {

		case <-finishCh:
			break watchdog

		case <-externalCtx.Done():
			break watchdog

		case workerError = <-errCh:
			shutdownWorkers()
			break watchdog

		case <-progressTick:
			curPct := 100 * atomic.LoadUint64(blocksDone) / blocksTotal
			if curPct != lastPct {
				lastPct = curPct
				fmt.Fprintf(os.Stderr, "%d%% of blocks written\r", lastPct)
			}
		}
	}

	wg.Wait()
	close(errCh) // closing a buffered channel keeps any buffered values for <-

	if workerError != nil {
		return workerError
	}
	if err := <-errCh; err != nil {
		return err
	}
	return externalCtx.Err()
}
