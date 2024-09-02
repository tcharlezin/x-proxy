package proxy

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
	"x-proxy/app"
)

type HandlerTransport struct {
}

func (HandlerTransport) RoundTrip(request *http.Request) (*http.Response, error) {

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
	}

	return transport.RoundTrip(request)
}

func Proxy(writer http.ResponseWriter, request *http.Request) {

	proxyRequest := request.WithContext(context.Background())
	reverseProxy, err := createReverseProxy()

	if err != nil {
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte("Reverse proxy not available!"))
		app.Application.Log.Error("createReverseProxy - ", err)
		return
	}

	app.Application.Log.Info("Forwarding request to reverse proxy")

	reverseProxy.ServeHTTP(writer, proxyRequest)
}

func modifyResponse(resp *http.Response) error {

	readBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Decode Gzip
	reader := bytes.NewReader(readBody)
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	output, err := io.ReadAll(gzreader)
	if err != nil {
		return err
	}

	htmlResponse := string(output)
	htmlResponse = strings.ReplaceAll(htmlResponse, "https://twitter.com", os.Getenv("TWITTER_HOST"))

	// Encode with GZip
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err = gz.Write([]byte(htmlResponse)); err != nil {
		return err
	}
	if err = gz.Close(); err != nil {
		return err
	}

	gzipEncoded := b.Bytes()
	resp.Body = io.NopCloser(bytes.NewReader(gzipEncoded))
	return nil
}

func reverseProxyErrorHandler(writer http.ResponseWriter, request *http.Request, err error) {

	if request.Body == nil {
		return
	}

	b, err := io.ReadAll(request.Body)

	if err != nil {
		app.Application.Log.Error("io.ReadAll - ", err)
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte("Reverse proxy not available!"))
		return
	}

	err = request.Body.Close()

	if err != nil {
		app.Application.Log.Error("request.Body.Close - ", err)
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte("Reverse proxy not available!"))
		return
	}

	body := io.NopCloser(bytes.NewReader(b))
	request.Body = body
}

func createReverseProxy() (*httputil.ReverseProxy, error) {

	director, err := getReverseProxyDirector()

	if err != nil {
		return nil, err
	}

	proxy := &httputil.ReverseProxy{
		Director:       director,
		Transport:      HandlerTransport{},
		ErrorHandler:   reverseProxyErrorHandler,
		ModifyResponse: modifyResponse,
	}

	return proxy, err
}

func getReverseProxyDirector() (func(request *http.Request), error) {
	var err error
	var hostAddress *url.URL

	hostAddress, err = url.Parse(app.Application.TargetHost)

	if err != nil {
		return nil, err
	}

	director := func(request *http.Request) {
		request.Host = hostAddress.Host
		request.URL.Scheme = hostAddress.Scheme
		request.URL.Host = hostAddress.Host
	}

	return director, nil
}
