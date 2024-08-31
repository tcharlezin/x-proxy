package app

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"x-proxy/app/setup"
)

type Config struct {
	WebPort    string
	TargetHost string
	Log        *slog.Logger
}

var Application = Config{
	WebPort:    os.Getenv("WEB_PORT"),
	TargetHost: os.Getenv("TARGET_HOST"),
	Log:        setup.SetupLog(),
}

func init() {

	if Application.WebPort == "" {
		log.Fatalln("WebPort not configured!")
	}

	if Application.TargetHost == "" {
		log.Fatalln("TargetHost not configured!")
	}

	http.DefaultTransport.(*http.Transport).MaxIdleConns = 1000
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1000
}
