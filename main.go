package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/golang/glog"
)

type stringMap map[string]string

var esEndpoint string
var listenHostAddr string
var awsRegion string
var hosts stringMap

// Value interface
func (m *stringMap) String() string {
	a := make([]string, 0, len(*m))
	for x := range *m {
		a = append(a, x)
	}
	return strings.Join(a, ",")
}

func (m *stringMap) Set(value string) error {
	s := strings.Split(value, "=")
	if len(s) != 2 {
		return errors.New("not of form src=dest")
	}
	src := s[0]
	dest := s[1]
	if !strings.Contains(dest, ".") {
		dest = fmt.Sprintf("%s.%s.es.amazonaws.com", dest, awsRegion)
	}
	(*m)[src] = dest
	return nil
}

func init() {
	flag.StringVar(&esEndpoint, "es", "", "elasticsearch endpoint")
	flag.Var(&hosts, "rewrite", "Map host to cluster, e.g. HOST=CLUSTERNAME. Can be repeated.")
	flag.StringVar(&listenHostAddr, "listen", ":9200", "endpoint on which to listen")

	flag.StringVar(&awsRegion, "region", "", "AWS Region")
}

func envToFlag(envName, flagName string) {
	v := os.Getenv(envName)
	if v != "" {
		flag.Set(flagName, v)
	}
}

func main() {
	flag.Set("logtostderr", "1")
	hosts = make(stringMap)

	flag.Parse()

	envToFlag("AWS_REGION", "region")
	envToFlag("ES", "es")
	envToFlag("LISTEN", "listen")
	envToFlag("GLOG_v", "v")

	if esEndpoint == "" {
		glog.Fatal("Elasticsearch endpoint (es) is required, even with vhosts (rewrite)")
	}
	if awsRegion == "" {
		glog.Fatal("AWS region flag (region) is required")
	}

	target, err := url.Parse(esEndpoint)
	if err != nil {
		glog.Fatalf("cannot parse es argument (%q) as URL", esEndpoint)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	credentials := defaults.CredChain(defaults.Config(), defaults.Handlers())

	signingRoundTripper := NewSigningRoundTripper(proxy.Transport, awsRegion, credentials, hosts)
	proxy.Transport = signingRoundTripper

	s := &http.Server{
		Addr:           listenHostAddr,
		Handler:        proxy,
		ReadTimeout:    120 * time.Second,
		WriteTimeout:   120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	glog.Infof("Listening on %s", listenHostAddr)
	err = s.ListenAndServe()
	glog.Fatalf("error listening on %q for http requests: %v", listenHostAddr, err)
}
