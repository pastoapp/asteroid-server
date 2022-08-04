package orbitdb

import (
	berty "berty.tech/go-orbit-db"
	"berty.tech/go-orbit-db/iface"
	"context"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	"log"
	"net/http"
)

// Client is the basic client from the berty library
var Client berty.OrbitDB

// init runs on module initialization
func init() {
	log.SetPrefix("[orbitdb/orbitdb] ")
}

// createUrlHttpApi creates a new http.Transport layer for the running IPFS instance. It's going to be used with Client.
func createUrlHttpApi(ipfsApiURL string) (*httpapi.HttpApi, error) {
	return httpapi.NewURLApiWithClient(ipfsApiURL, &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
	})
}

// InitializeOrbitDB initializes a new ODB instance; taking the IPFS-Node API and the store directory into account.
func InitializeOrbitDB(ipfsApiURL, orbitDbDirectory string) (context.CancelFunc, error) {
	// A production version could also take more HTTP-API and ODB config options into account.
	ctx, cancel := context.WithCancel(context.Background())
	odb, err := NewOrbitDB(ctx, orbitDbDirectory, ipfsApiURL)
	if err != nil {
		log.Print(err)
		cancel()
		return nil, err
	}
	Client = odb
	return cancel, nil
}

// NewOrbitDB creates a OrbitDB instance in memory and returns a reference.
func NewOrbitDB(ctx context.Context, dbPath, ipfsApiURL string) (iface.OrbitDB, error) {
	coreAPI, err := createUrlHttpApi(ipfsApiURL)

	if err != nil {
		log.Printf("Error creating Core API: %v", err)
		return nil, err
	}

	options := &berty.NewOrbitDBOptions{
		Directory: &dbPath,
	}

	return berty.NewOrbitDB(ctx, coreAPI, options)
}
