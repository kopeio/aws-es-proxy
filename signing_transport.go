package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/private/signer/v4"
	"github.com/golang/glog"
)

const SERVICE_NAME = "es"

type SigningRoundTripper struct {
	region      string
	inner       http.RoundTripper
	credentials *credentials.Credentials
}

var _ http.RoundTripper = &SigningRoundTripper{}

func NewSigningRoundTripper(inner http.RoundTripper, region string, credentials *credentials.Credentials) *SigningRoundTripper {
	if inner == nil {
		inner = http.DefaultTransport
	}
	p := &SigningRoundTripper{inner: inner, region: region, credentials: credentials}
	return p
}

func (p *SigningRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	glog.V(2).Infof("Got request: %s %s", req.Method, req.URL)

	// Fix the host header in case broken by proxy-rewrite
	if req.URL.Host != "" {
		req.Host = req.URL.Host
	}

	// I think the AWS authentication proxy does not like forwarded headers
	for k := range req.Header {
		lk := strings.ToLower(k)
		if lk == "x-forwarded-host" {
			delete(req.Header, k)
		}
		if lk == "x-forwarded-for" {
			delete(req.Header, k)
		}
		if lk == "x-forwarded-proto" {
			delete(req.Header, k)
		}
		if lk == "x-forward-for" {
			delete(req.Header, k)
		}
		if lk == "x-forward-proto" {
			delete(req.Header, k)
		}
		if lk == "x-forward-port" {
			delete(req.Header, k)
		}
	}

	// We're going to put our own auth headers on here
	delete(req.Header, "Authorization")

	var body []byte
	var err error

	if req.Body != nil {
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			glog.Infof("error reading request body: %v", err)
			return nil, err
		}
	}

	if req.Method == "GET" || req.Method == "HEAD" {
		delete(req.Header, "Content-Length")
	}

	oldPath := req.URL.Path
	if oldPath != "" {
		// Escape the path before signing so that the path in the signature and
		// the path in the request match.
		req.URL.Path = req.URL.EscapedPath()
		glog.V(4).Infof("Path -> %q", req.URL.Path)
	}

	awsReq := &request.Request{}
	awsReq.Config.Credentials = p.credentials
	awsReq.Config.Region = aws.String(p.region)
	awsReq.ClientInfo.ServiceName = SERVICE_NAME
	awsReq.HTTPRequest = req
	awsReq.Time = time.Now()
	awsReq.ExpireTime = 0
	if body != nil {
		awsReq.Body = bytes.NewReader(body)
	}

	if glog.V(4) {
		awsReq.Config.LogLevel = aws.LogLevel(aws.LogDebugWithSigning)
		awsReq.Config.Logger = aws.NewDefaultLogger()
	}

	v4.Sign(awsReq)

	if awsReq.Error != nil {
		glog.Warningf("error signing request: %v", awsReq.Error)
		return nil, awsReq.Error
	}

	req.URL.Path = oldPath

	if body != nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	response, err := p.inner.RoundTrip(req)

	if err != nil {
		glog.Warning("Request error: ", err)
		return nil, err
	} else {
		glog.V(2).Infof("response %s", response.Status)
		return response, err
	}
}
