package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"x-proxy/app"
)

type DebugTransport struct{}

func (DebugTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return http.DefaultTransport.RoundTrip(request)
}

func Proxy(writer http.ResponseWriter, request *http.Request) {

	proxyRequest := request.WithContext(context.TODO())
	reverseProxy, err := createReverseProxy()

	if err != nil {
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte("Reverse proxy not available!"))
		app.Application.Log.Error("createReverseProxy - ", err)
		return
	}

	app.Application.Log.Info("Forwarding request to reverse proxy")

	reverseProxy.ModifyResponse = rewriteBody
	reverseProxy.ServeHTTP(writer, proxyRequest)
}

func rewriteBody(resp *http.Response) (err error) {

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	body := io.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return nil
}

func reverseProxyErrorHandler(writer http.ResponseWriter, request *http.Request, err error) {
	app.Application.Log.Error("ReverseProxyErrorHandler: ", err)

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

	var result map[string]interface{}
	err = json.Unmarshal(b, &result)

	if err != nil {
		app.Application.Log.Error("json.Unmarshal - ", err)
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte("Reverse proxy not available!"))
		return
	}

	writer.WriteHeader(http.StatusBadGateway)
	_, _ = writer.Write([]byte("Reverse proxy not available!"))
}

func createReverseProxy() (*httputil.ReverseProxy, error) {

	director, err := getReverseProxyDirector()

	if err != nil {
		return nil, err
	}

	proxy := &httputil.ReverseProxy{
		Director:     director,
		Transport:    DebugTransport{},
		ErrorHandler: reverseProxyErrorHandler,
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
