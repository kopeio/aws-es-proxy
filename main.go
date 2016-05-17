package main

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/golang/glog"
)

var esEndpoint string
var listenHostAddr string
var awsRegion string

func init() {
	flag.StringVar(&esEndpoint, "es", "", "elasticsearch endpoint")
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

	flag.Parse()

	envToFlag("AWS_REGION", "region")
	envToFlag("ES", "es")
	envToFlag("LISTEN", "listen")
	envToFlag("GLOG_v", "v")

	if esEndpoint == "" {
		glog.Fatal("elasticsearch endpoint flag (es) is required")
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

	signingRoundTripper := NewSigningRoundTripper(proxy.Transport, awsRegion, credentials)
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
