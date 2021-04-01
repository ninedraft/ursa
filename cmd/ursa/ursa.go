package main

import (
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ninedraft/gemax/gemax"
	"github.com/ninedraft/gemax/gemax/status"
)

func main() {
	var exitCode int
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	var certFilename string
	flag.StringVar(&certFilename, "file-cert", "cert.crt", `certfile`)
	var keyFilename string
	flag.StringVar(&keyFilename, "file-key", "cert.key", `keyfile`)
	flag.Parse()

	var cert, errCert = tls.LoadX509KeyPair(certFilename, keyFilename)
	if errCert != nil {
		log.Println(errCert)
		exitCode = 1
		return
	}

	var server = &gemax.Server{
		Addr:    "localhost:9999",
		Logf:    log.Printf,
		Handler: routes,
	}
	var ctx = context.Background()
	var errServe = server.ListenAndServe(ctx, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	})
	switch {
	case errServe == nil ||
		errors.Is(errServe, context.Canceled):
	// ok
	default:
		log.Println("serving: ", errServe)
		exitCode = 1
	}
}

//go:embed index.gmi
var indexPageData []byte
var indexPage = gemax.ServeContent("text/gemini", indexPageData)

func routes(ctx context.Context, rw gemax.ResponseWriter, req gemax.IncomingRequest) {
	var path = req.URL().Path
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	switch {
	case path == "/index":
		indexPage(ctx, rw, req)
	case path == "" || path == "/":
		gemax.Redirect(rw, req, "/index", status.Redirect)
	case strings.HasPrefix(path, "/ipfs/"):
	case strings.HasPrefix(path, "/ipns/"):
	case strings.HasPrefix(path, "/fetch/hash"):
		rw.WriteStatus(status.Input, "what object to load?")
	case strings.HasPrefix(path, "/fetch/name"):
		rw.WriteStatus(status.Input, "what name to resolve?")
	default:
		gemax.NotFound(rw, req)
	}
}
