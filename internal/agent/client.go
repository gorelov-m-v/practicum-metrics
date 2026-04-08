// Package agent implements metrics collection and sending for the monitoring agent.
package agent

import "github.com/go-resty/resty/v2"

type HTTPClient interface {
	Post(url string) (*resty.Response, error)
}

type RestyClientAdapter struct {
	client *resty.Client
}

func NewRestyClientAdapter(client *resty.Client) *RestyClientAdapter {
	return &RestyClientAdapter{
		client: client,
	}
}

func (r *RestyClientAdapter) Post(url string) (*resty.Response, error) {
	return r.client.R().Post(url)
}
