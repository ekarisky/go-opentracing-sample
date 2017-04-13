package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ekarisky/go-opentracing-sample/hello"
	"github.com/opentracing/opentracing-go"
	grace "gopkg.in/paytm/grace.v1"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

func main() {
	log.Println("starting tracer on")
	go setupTracer(8700, 3600, "")

	log.Println("app started")

	hello := hello.NewHelloModule()
	hello.InitHandlers()

	log.Fatal(grace.Serve(":9000", nil))
}

// setupTracer must be called in a separate goroutine, as it's blocking
func setupTracer(appdashPort int, ttl int, server string) {

	time.Sleep(5) // sleep 5 seconds so we don't run into port conflicts

	// Tracer setup
	memStore := appdash.NewMemoryStore()

	// keep last hour of traces.
	store := &appdash.RecentStore{
		MinEvictAge: time.Duration(ttl) * time.Second,
		DeleteStore: memStore,
	}
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		log.Fatalln("appdash", err)
	}

	collectorPort := l.Addr().String()
	log.Println("collector listening on", collectorPort)

	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
	go cs.Start()

	if server == "" {
		server = fmt.Sprintf("http://localhost:%d", appdashPort)
	}

	appdashURL, err := url.Parse(server)
	tapp, err := traceapp.New(nil, appdashURL)
	if err != nil {
		log.Fatal(err)
	}
	tapp.Store = store
	tapp.Queryer = memStore

	tracer := appdashot.NewTracer(appdash.NewRemoteCollector(collectorPort))
	opentracing.InitGlobalTracer(tracer)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appdashPort), tapp))
}
