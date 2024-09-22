package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/tgragnato/magnetico/dht"
	"github.com/tgragnato/magnetico/dht/mainline"
	"github.com/tgragnato/magnetico/metadata"
)

const VERSION = "1.0.0"

var opFlags struct {
	ImportSourceName string
	ImportURL        string
	ImportDebug      bool

	IndexerAddrs        []string
	IndexerMaxNeighbors uint

	LeechMaxN          int
	BootstrappingNodes []string
	FilterNodesCIDRs   []net.IPNet
}

func main() {
	// opFlags is the "operational flags"
	if parseFlags() != nil {
		// Do not print any error messages as jessevdk/go-flags already did.
		return
	}

	// Handle Ctrl-C gracefully.
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	trawlingManager := dht.NewManager(opFlags.IndexerAddrs, opFlags.IndexerMaxNeighbors, opFlags.BootstrappingNodes, opFlags.FilterNodesCIDRs)
	metadataSink := metadata.NewSink(5*time.Second, opFlags.LeechMaxN, opFlags.FilterNodesCIDRs)

	log.Printf("Running Magnetico-Bitmagnet V%s\n", VERSION)

	// The Event Loop
	for stopped := false; !stopped; {
		select {
		case result := <-trawlingManager.Output():
			metadataSink.Sink(result)

		case md := <-metadataSink.Drain():
			InfoHashHex := hex.EncodeToString(md.InfoHash)
			t := time.Unix(md.DiscoveredOn, 0).UTC()
			FormattedTime := t.Format(time.RFC3339)

			data := map[string]interface{}{
				"infoHash":    InfoHashHex,
				"name":        md.Name,
				"size":        md.TotalSize,
				"publishedAt": FormattedTime,
				"source":      opFlags.ImportSourceName,
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Println("Error marshalling JSON:", err)
				break
			}

			dataBuffer := bytes.NewBuffer(jsonData)
			dataBuffer.Write([]byte("\n"))
			resp, err := http.Post(opFlags.ImportURL, "application/json", dataBuffer)
			if err != nil {
				fmt.Println("Error sending POST request:", err)
				break
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				break
			}
			if opFlags.ImportDebug == true {
				log.Printf("Response: %s\n", string(body))
			}

		case <-interruptChan:
			trawlingManager.Terminate()
			stopped = true
		}
	}
}

func parseFlags() error {
	var cmdF struct {
		Version bool `long:"version" description:"Print the version and exit."`

		ImportSourceName string `long:"import-source" description:"The name to use as \"source\" when importing to bitmagnet." default:"magnetico"`
		ImportURL        string `long:"import-url" description:"The URL to use when importing to bitmagnet." default:"http://localhost:3333/import"`
		ImportDebug      bool   `long:"import-debug" description:"Print the responses of the POST requests to bitmagnet when importing."`

		IndexerAddrs        []string `long:"indexer-addr" description:"Address(es) to be used by indexing DHT nodes." default:"0.0.0.0:0"`
		IndexerMaxNeighbors uint     `long:"indexer-max-neighbors" description:"Maximum number of neighbors of an indexer." default:"5000"`

		LeechMaxN uint `long:"leech-max-n" description:"Maximum number of leeches." default:"1000"`
		MaxRPS    uint `long:"max-rps" description:"Maximum requests per second." default:"500"`

		BootstrappingNodes []string `long:"bootstrap-node" description:"Host(s) to be used for bootstrapping." default:"dht.tgragnato.it"`
		FilterNodesCIDRs   []string `long:"filter-nodes-cidrs" description:"List of CIDRs on which Magnetico can operate. Empty is open mode." default:""`
	}

	if _, err := flags.Parse(&cmdF); err != nil {
		return err
	}

	if cmdF.Version {
		fmt.Printf("Magnetico-Bitmagnet V%s\n", VERSION)
		os.Exit(0)
	}

	opFlags.ImportSourceName = cmdF.ImportSourceName
	opFlags.ImportURL = cmdF.ImportURL
	opFlags.ImportDebug = cmdF.ImportDebug

	if err := checkAddrs(cmdF.IndexerAddrs); err != nil {
		log.Fatalf("Of argument (list) `trawler-ml-addr` %s\n", err.Error())
	} else {
		opFlags.IndexerAddrs = cmdF.IndexerAddrs
	}

	opFlags.IndexerMaxNeighbors = cmdF.IndexerMaxNeighbors

	opFlags.LeechMaxN = int(cmdF.LeechMaxN)
	if opFlags.LeechMaxN > 1000 {
		log.Println(
			"Beware that on many systems max # of file descriptors per process is limited to 1024. " +
				"Setting maximum number of leeches greater than 1k might cause \"too many open files\" errors!",
		)
	}

	mainline.DefaultThrottleRate = int(cmdF.MaxRPS)
	opFlags.BootstrappingNodes = cmdF.BootstrappingNodes

	opFlags.FilterNodesCIDRs = []net.IPNet{}
	for _, cidr := range cmdF.FilterNodesCIDRs {
		if cidr == "" {
			continue
		}
		if _, ipnet, err := net.ParseCIDR(cidr); err == nil {
			opFlags.FilterNodesCIDRs = append(opFlags.FilterNodesCIDRs, *ipnet)
		} else {
			log.Fatalf("Error while parsing CIDR %s: %s\n", cidr, err.Error())
		}
	}
	if len(opFlags.FilterNodesCIDRs) != 0 && reflect.DeepEqual(cmdF.BootstrappingNodes, []string{"dht.tgragnato.it"}) {
		log.Fatalln("You should specify your own internal bootstrapping nodes in filter mode.")
	}

	return nil
}

func checkAddrs(addrs []string) error {
	for _, addr := range addrs {
		// We are using ResolveUDPAddr but it works equally well for checking TCPAddr(esses) as
		// well.
		_, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return err
		}
	}
	return nil
}
