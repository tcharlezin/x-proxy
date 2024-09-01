package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"x-proxy/app"
)

type DebugTransport struct{}

func (DebugTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)

	if err != nil {
		app.Application.Log.Info("RoundTrip Error: ", err)
	}

	return response, err
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

func modifyResponse(response *http.Response) error {
	app.Application.Log.Info("ModifyResponse Status: ", response.StatusCode)
	for key, value := range response.Header {
		app.Application.Log.Info("Key: ", key, "Value: ", value)
	}
	return nil
}

func reverseProxyErrorHandler(writer http.ResponseWriter, request *http.Request, err error) {
	app.Application.Log.Error("ReverseProxyErrorHandler: ", "error: ", err)
	app.Application.Log.Error("RequestURI: ", request.RequestURI)
	app.Application.Log.Error("Method: ", request.Method)
	app.Application.Log.Error("URL: ", request.URL)

	// TODO: improve logs
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
		Transport:      DebugTransport{},
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
