package utils

import (
	"context"
	"net"
	"net/http"
	"time"
)

type HttpClient struct {
	TimeOut               time.Duration `yaml:"TimeOut"`
	MaxIdleConnsPerHost   int           `yaml:"MaxIdleConnsPerHost"`
	MaxConnsPerHost       int           `yaml:"MaxConnsPerHost"`
	IdleConnTimeout       time.Duration `yaml:"IdleConnTimeout"`
	TLSHandshakeTimeout   time.Duration `yaml:"TLSHandshakeTimeout"`
	ExpectContinueTimeout time.Duration `yaml:"ExpectContinueTimeout"`
	TcpTimeOut            time.Duration `yaml:"TcpTimeOut"`
	TcpKeepAlive          time.Duration `yaml:"TcpKeepAlive"`
}

var Client *http.Client

func InitHttpClient(hc HttpClient) {
	Client = &http.Client{
		Timeout: hc.TimeOut,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout:   hc.TcpTimeOut,
					KeepAlive: hc.TcpKeepAlive,
				}
				return dialer.DialContext(ctx, network, addr)
			},
			MaxIdleConnsPerHost: hc.MaxIdleConnsPerHost,
			//MaxConnsPerHost:       hc.MaxConnsPerHost,
			IdleConnTimeout:       hc.IdleConnTimeout,
			TLSHandshakeTimeout:   hc.TLSHandshakeTimeout,
			ExpectContinueTimeout: hc.ExpectContinueTimeout,
		},
	}
}
