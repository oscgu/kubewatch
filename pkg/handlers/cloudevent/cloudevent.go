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

package cloudevent

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"

	"time"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"k8s.io/apimachinery/pkg/runtime"
)

var cloudEventErrMsg = `
%s

You need to set Cloudevents webhook url
using "--url/-u" or using environment variables:

export KW_CLOUDEVENT_URL=webhook_url

Command line flags will override environment variables

`

// Webhook handler implements handler.Handler interface,
// Notify event to Webhook channel
type CloudEvent struct {
	Url       string
	StartTime uint64
	Counter   atomic.Uint64
	Client    *fasthttp.Client
}

type CloudEventMessage struct {
	SpecVersion     string                `json:"specversion"`
	Type            string                `json:"type"`
	Source          string                `json:"source"`
	Subject         string                `json:"subject"`
	ID              string                `json:"id"`
	Time            string                `json:"time"`
	DataContentType string                `json:"datacontenttype"`
	Data            CloudEventMessageData `json:"data"`
}

// EventMeta containes the meta data about the event occurred
type CloudEventMessageData struct {
	Operation   string         `json:"operation"`
	Kind        string         `json:"kind"`
	ClusterUid  string         `json:"clusterUid"`
	Description string         `json:"description"`
	ApiVersion  string         `json:"apiVersion"`
	Obj         runtime.Object `json:"obj"`
	OldObj      runtime.Object `json:"oldObj"`
}

func (m *CloudEvent) Init(c *config.Config) error {
	m.Url = c.Handler.CloudEvent.Url
	m.StartTime = uint64(time.Now().Unix())
	m.Counter.Store(0)
	m.Client = &fasthttp.Client{}

	if m.Url == "" {
		m.Url = os.Getenv("KW_CLOUDEVENT_URL")
	}

	if m.Url == "" {
		return fmt.Errorf(cloudEventErrMsg, "Missing cloudevent url")
	}

	return nil
}

func (m *CloudEvent) Handle(e event.Event) {
	m.Counter.Add(1)
	message := m.prepareMessage(e)

	err := m.postMessage(message)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	log.Printf("Message successfully sent to %s", m.Url)
}

func (m *CloudEvent) prepareMessage(e event.Event) *CloudEventMessage {
	return &CloudEventMessage{
		SpecVersion:     "1.0",
		Type:            "KUBERNETES_TOPOLOGY_CHANGE",
		Source:          "https://github.com/aantn/kubewatch",
		ID:              fmt.Sprintf("%v-%v", m.StartTime, m.Counter.Load()),
		Time:            time.Now().Format(time.RFC3339), // https://github.com/cloudevents/spec/blob/main/cloudevents/spec.md#time
		DataContentType: "application/json",
		Data: CloudEventMessageData{
			Operation:   m.formatReason(e),
			Kind:        e.Kind,
			ApiVersion:  e.ApiVersion,
			ClusterUid:  "TODO",
			Description: e.Message(),
			Obj:         e.Obj,
			OldObj:      e.OldObj,
		},
	}
}

func (m *CloudEvent) formatReason(e event.Event) string {
	switch e.Reason {
	case "Created":
		return "create"
	case "Updated":
		return "update"
	case "Deleted":
		return "delete"
	default:
		return "unknown"
	}
}

func (m *CloudEvent) postMessage(webhookMessage *CloudEventMessage) error {
	message, err := json.Marshal(webhookMessage)
	if err != nil {
		return err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(m.Url)
	req.SetBody(message)
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	if err := m.Client.Do(req, res); err != nil {
		return err
	}

	return nil
}
