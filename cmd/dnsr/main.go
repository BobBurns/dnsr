package main

import (
	"flag"
	"os"
	"sync"
	"time"

	"code.google.com/p/go.net/idna"
	"github.com/domainr/dnsr"
	"github.com/miekg/dns"
	"github.com/wsxiaoys/terminal/color"
)

const (
	timeout = 2000 * time.Millisecond
)

var (
	verbose  bool
	resolver = dnsr.New(10000)
)

func init() {
	flag.BoolVar(
		&verbose,
		"v",
		false,
		"print verbose info to the console",
	)
}

func logV(fmt string, args ...interface{}) {
	if !verbose {
		return
	}
	color.Printf(fmt, args...)
}

func main() {
	flag.Usage = func() {
		color.Fprintf(os.Stderr, "Usage: %s [arguments] <name> [type]\n\nAvailable arguments:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	qtype := ""
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
	} else if _, isType := dns.StringToType[args[len(args)-1]]; len(args) > 1 && isType {
		qtype, args = args[len(args)-1], args[:len(args)-1]
	}
	var wg sync.WaitGroup
	start := time.Now()
	for _, name := range args {
		wg.Add(1)
		go func(name string, qtype string) {
			query(name, qtype)
			wg.Done()
		}(name, qtype)
	}
	wg.Wait()
	logV("\n@{w};; Total elapsed: %s\n", time.Since(start).String())
}

func query(name, qtype string) {
	start := time.Now()
	qname, err := idna.ToASCII(name)
	if err != nil {
		color.Fprintf(os.Stderr, "Invalid IDN domain name: %s\n", name)
		os.Exit(1)
	}

	rrc := resolver.Resolve(qname, qtype)
	rrs := []*dnsr.RR{}
	for rr := range rrc {
		if rr == nil {
			logV("@{r}\n;; NIL RR!\n")
			continue
		}
		rrs = append(rrs, rr)
	}

	color.Printf("\n")
	if len(rrs) > 0 {
		color.Printf("@{g};; RESULTS:\n")
	}
	for _, rr := range rrs {
		color.Printf("@{g}%s\n", rr.String())
	}

	if rrs == nil {
		color.Printf("@{y};; NIL\t%s\t%s\n", name, qtype)
	} else if len(rrs) > 0 {
		color.Printf("@{g};; TRUE\t%s\t%s\n", name, qtype)
	} else {
		color.Printf("@{r};; FALSE\t%s\t%s\n", name, qtype)
	}

	logV("@{.w};; Elapsed: %s\n", time.Since(start).String())
}
