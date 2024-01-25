// Copyright © 2024 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pulsar

//go:generate paramgen -output=paramgen_src.go SourceConfig

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/google/uuid"
)

type Source struct {
	sdk.UnimplementedSource

	consumer pulsar.Consumer
	received map[string]pulsar.Message
	mx       *sync.Mutex
	config   SourceConfig
}

func NewSource() sdk.Source {
	source := &Source{
		mx:       &sync.Mutex{},
		received: make(map[string]pulsar.Message),
	}

	return sdk.SourceWithMiddleware(source, sdk.DefaultSourceMiddleware()...)
}

func (s *Source) Parameters() map[string]sdk.Parameter {
	return s.config.Parameters()
}

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	sdk.Logger(ctx).Info().Msg("Configuring Source...")

	if err := sdk.Util.ParseConfig(cfg, &s.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

func (s *Source) Open(ctx context.Context, pos sdk.Position) error {
	var logger log.Logger
	if s.config.DisableLogging {
		logger = log.DefaultNopLogger()
	}

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                        s.config.URL,
		ConnectionTimeout:          s.config.ConnectionTimeout,
		OperationTimeout:           s.config.OperationTimeout,
		MaxConnectionsPerBroker:    s.config.MaxConnectionsPerBroker,
		MemoryLimitBytes:           s.config.MemoryLimitBytes,
		EnableTransaction:          s.config.EnableTransaction,
		TLSKeyFilePath:             s.config.TLSKeyFilePath,
		TLSCertificateFile:         s.config.TLSCertificateFile,
		TLSTrustCertsFilePath:      s.config.TLSTrustCertsFilePath,
		TLSAllowInsecureConnection: s.config.TLSAllowInsecureConnection,
		TLSValidateHostname:        s.config.TLSValidateHostname,

		Logger: logger,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	s.consumer, err = client.Subscribe(pulsar.ConsumerOptions{
		Topic:            s.config.Topic,
		SubscriptionName: s.config.SubscriptionName,
		Type:             pulsar.Exclusive,
	})
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	if pos != nil {
		p, err := parsePosition(pos)
		if err != nil {
			return err
		}

		if s.config.SubscriptionName != "" && s.config.SubscriptionName != p.SubscriptionName {
			return fmt.Errorf("the old position contains a different subscription name than the connector configuration (%q vs %q), please check if the configured subscription name changed since the last run", p.SubscriptionName, s.config.SubscriptionName)
		}

		s.config.SubscriptionName = p.SubscriptionName
	}

	if s.config.SubscriptionName == "" {
		// this must be the first run of the connector, create a new group ID
		s.config.SubscriptionName = uuid.NewString()
		sdk.Logger(ctx).Info().Str("subscriptionName", s.config.SubscriptionName).Msg("assigning source to new subscription")
	}

	return nil
}

const (
	// MetadataPulsarTopic is the metadata key for storing the pulsar topic
	MetadataPulsarTopic = "pulsar.topic"
)

func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	sdk.Logger(ctx).Debug().Msg("reading message")
	msg, err := s.consumer.Receive(ctx)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("failed to receive message: %w", err)
	}

	s.mx.Lock()
	s.received[msg.ID().String()] = msg
	s.mx.Unlock()

	position := Position{s.config.SubscriptionName}
	sdkPos := position.ToSDKPosition()

	metadata := sdk.Metadata{MetadataPulsarTopic: msg.Topic()}
	metadata.SetCreatedAt(msg.EventTime())

	key := sdk.RawData(msg.Key())
	payload := sdk.RawData(msg.Payload())

	newRecord := sdk.Util.Source.NewRecordCreate(sdkPos, metadata, key, payload)

	return newRecord, nil
}

func (s *Source) Ack(ctx context.Context, position sdk.Position) error {
	sdk.Logger(ctx).Debug().Str("MessageID", string(position)).Msg("Attempting to ack message")

	msgID, err := pulsar.DeserializeMessageID(position)
	if err != nil {
		return fmt.Errorf("failed to deserialize message ID: %w", err)
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	msg, ok := s.received[msgID.String()]
	if ok {
		delete(s.received, msgID.String())
		return s.consumer.Ack(msg)
	}

	return fmt.Errorf("message not found for position: %s", string(position))
}

func (s *Source) Teardown(_ context.Context) error {
	if s.consumer != nil {
		s.consumer.Close()
	}
	return nil
}

type Position struct {
	SubscriptionName string `json:"subscriptionName"`
}

func parsePosition(pos sdk.Position) (Position, error) {
	var p Position
	err := json.Unmarshal(pos, &p)

	return p, err
}

func (p Position) ToSDKPosition() sdk.Position {
	bs, err := json.Marshal(p)
	if err != nil {
		// this error should not be possible
		panic(fmt.Errorf("error marshaling position to JSON: %w", err))
	}

	return sdk.Position(bs)
}
