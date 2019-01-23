package templates

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	tmplHttpClient *http.Client
)

func init() {
	// interface to use for http requests, if not specified, will not be enabled
	// example: 10.0.0.77
	interfaceToUse := os.Getenv("YAGPDB_TMPL_HTTP_INTERFACE_ADDRESS")
	if interfaceToUse == "" {
		return
	}

	// resolve the tcp adress
	ipAddr, err := net.ResolveTCPAddr("tcp", interfaceToUse+":0")
	if err != nil {
		panic(err)
	}

	// create the dialer
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
		LocalAddr: ipAddr,
	}

	// set up the slightly modified http transport
	roundTripper := new(http.Transport)
	*roundTripper = *(http.DefaultTransport.(*http.Transport))

	roundTripper.DialContext = dialer.DialContext
	roundTripper.Dial = nil

	// set up the custom client
	c := *http.DefaultClient
	tmplHttpClient = &c
	tmplHttpClient.Transport = roundTripper
}

type RequestReponse struct {
	Headers    http.Header
	StatusCode int
	Status     string

	Body []byte
}

func (rr *RequestReponse) StringBody() string {
	return string(rr.Body)
}

func (c *Context) tmplHTTPGet(url string, headers ...string) (interface{}, error) {
	if tmplHttpClient == nil {
		return nil, errors.New("HTTP disabled, YAGPDB_TMPL_HTTP_INTERFACE_ADDRESS not set, contact bot admin")
	}

	if c.IncreaseCheckCallCounter("http_requests", 3) {
		return nil, ErrTooManyCalls
	}

	if len(headers)%2 != 0 {
		return nil, errors.New("invalid headers, not dividable by 2, supply key value pairs.")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(headers); i += 2 {
		k := headers[i]
		v := headers[i+1]

		req.Header.Set(k, v)
	}

	resp, err := tmplHttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// read max 1 megabyte of the body
	r := io.LimitReader(resp.Body, 1000000)
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &RequestReponse{
		Headers:    resp.Header,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,

		Body: b,
	}, nil
}
