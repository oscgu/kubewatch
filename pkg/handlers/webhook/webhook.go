/*
Copyright 2018 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"encoding/json"
	"time"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

var webhookErrMsg = `
%s

You need to set Webhook url, and Webhook cert if you use self signed certificates,
using "--url/-u" and "--cert", or using environment variables:

export KW_WEBHOOK_URL=webhook_url
export KW_WEBHOOK_CERT=/path/of/cert

Command line flags will override environment variables

`

// Webhook handler implements handler.Handler interface,
// Notify event to Webhook channel
type Webhook struct {
	Url    string
	Client *fasthttp.Client
}

// WebhookMessage for messages
type WebhookMessage struct {
	EventMeta EventMeta `json:"eventmeta"`
	Text      string    `json:"text"`
	Time      time.Time `json:"time"`
}

// EventMeta containes the meta data about the event occurred
type EventMeta struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Reason    string `json:"reason"`
}

// Init prepares Webhook configuration
func (m *Webhook) Init(c *config.Config) error {
	url := c.Handler.Webhook.Url
	cert := c.Handler.Webhook.Cert
	tlsSkip := c.Handler.Webhook.TlsSkip

	if url == "" {
		url = os.Getenv("KW_WEBHOOK_URL")
	}
	if cert == "" {
		cert = os.Getenv("KW_WEBHOOK_CERT")
	}

	m.Url = url
	m.Client = &fasthttp.Client{}

	if tlsSkip {
		m.Client.TLSConfig.InsecureSkipVerify = true
	} else {
		if cert == "" {
			log.Printf("No webhook cert is given")
		} else {
			caCert, err := os.ReadFile(cert)
			if err != nil {
				log.Printf("%s\n", err)
				return err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			m.Client.TLSConfig = &tls.Config{RootCAs: caCertPool}
		}

	}

	return checkMissingWebhookVars(m)
}

// Handle handles an event.
func (m *Webhook) Handle(e event.Event) {
	webhookMessage := prepareWebhookMessage(e, m)

	err := postMessage(m.Url, m.Client, webhookMessage)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	log.Printf("Message successfully sent to %s at %s ", m.Url, time.Now())
}

func checkMissingWebhookVars(s *Webhook) error {
	if s.Url == "" {
		return fmt.Errorf(webhookErrMsg, "Missing Webhook url")
	}

	return nil
}

func prepareWebhookMessage(e event.Event, m *Webhook) *WebhookMessage {
	return &WebhookMessage{
		EventMeta: EventMeta{
			Kind:      e.Kind,
			Name:      e.Name,
			Namespace: e.Namespace,
			Reason:    e.Reason,
		},
		Text: e.Message(),
		Time: time.Now(),
	}
}

func postMessage(url string, client *fasthttp.Client, webhookMessage *WebhookMessage) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)

	message, err := json.Marshal(webhookMessage)
	if err != nil {
		return err
	}
	req.SetBody(message)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")

	res := fasthttp.AcquireResponse()
	if err := client.Do(req, res); err != nil {
		return err
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)

	return nil
}
