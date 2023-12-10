package apachepulsar

//go:generate paramgen -output=paramgen_src.go SourceConfig

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/apache/pulsar-client-go/pulsar"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

type Source struct {
	sdk.UnimplementedSource

	consumer pulsar.Consumer
	received []pulsar.Message
	mx       *sync.Mutex

	config SourceConfig
}

type SourceConfig struct {
	Config

	URL              string           `json:"URL" validate:"required"`
	Topic            string           `json:"topic" validate:"required"`
	SubscriptionName string           `json:"subscriptionName" validate:"required"`
	SubscriptionType SubscriptionType `json:"subscriptionType"`
}

type SubscriptionType string

const (
	// Exclusive there can be only 1 consumer on the same topic with the same subscription name
	Exclusive SubscriptionType = "exclusive"

	// Shared subscription mode, multiple consumer will be able to use the same subscription name
	// and the messages will be dispatched according to
	// a round-robin rotation between the connected consumers
	Shared SubscriptionType = "shared"

	// Failover subscription mode, multiple consumer will be able to use the same subscription name
	// but only 1 consumer will receive the messages.
	// If that consumer disconnects, one of the other connected consumers will start receiving messages.
	Failover SubscriptionType = "failover"

	// KeyShared subscription mode, multiple consumer will be able to use the same
	// subscription and all messages with the same key will be dispatched to only one consumer
	KeyShared SubscriptionType = "key_shared"
)

func ParseSubscriptionType(s string) (SubscriptionType, bool) {
	switch s {
	case string(Exclusive):
		return Exclusive, true
	case string(Shared):
		return Shared, true
	case string(Failover):
		return Failover, true
	case string(KeyShared):
		return KeyShared, true
	default:
		return "", false
	}
}

func (s SubscriptionType) PulsarType() pulsar.SubscriptionType {
	switch s {
	case Exclusive:
		return pulsar.Exclusive
	case Shared:
		return pulsar.Shared
	case Failover:
		return pulsar.Failover
	case KeyShared:
		return pulsar.KeyShared
	default:
		return pulsar.Exclusive
	}
}

func NewSource() sdk.Source {
	return sdk.SourceWithMiddleware(&Source{mx: &sync.Mutex{}}, sdk.DefaultSourceMiddleware()...)
}

func (s *Source) Parameters() map[string]sdk.Parameter {
	return s.config.Parameters()
}

func (s *Source) Configure(ctx context.Context, cfg map[string]string) error {
	sdk.Logger(ctx).Info().Msg("Configuring Source...")
	err := sdk.Util.ParseConfig(cfg, &s.config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if stype, ok := cfg["subscriptionType"]; ok {
		subscriptionType, ok := ParseSubscriptionType(stype)
		if !ok {
			return fmt.Errorf("invalid subscriptionType: %s", stype)
		}

		s.config.SubscriptionType = subscriptionType
	}

	return nil
}

func (s *Source) Open(ctx context.Context, pos sdk.Position) error {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: s.config.URL,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	subscriptionType := s.config.SubscriptionType.PulsarType()

	s.consumer, err = client.Subscribe(pulsar.ConsumerOptions{
		Topic:            s.config.Topic,
		SubscriptionName: s.config.SubscriptionName,
		Type:             subscriptionType,
	})
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	return nil
}

func (s *Source) Read(ctx context.Context) (sdk.Record, error) {
	msg, err := s.consumer.Receive(ctx)
	if err != nil {
		return sdk.Record{}, fmt.Errorf("failed to receive message: %w", err)
	}

	s.mx.Lock()
	s.received = append(s.received, msg)
	s.mx.Unlock()

	var position sdk.Position
	var metadata sdk.Metadata
	var key sdk.Data
	var payload sdk.Data = sdk.RawData(msg.Payload())

	return sdk.Util.Source.NewRecordCreate(
		position,
		metadata,
		key,
		payload,
	), nil
}

func (s *Source) Ack(ctx context.Context, _ sdk.Position) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	var err error
	for _, msg := range s.received {
		ackErr := s.consumer.Ack(msg)
		if ackErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to ack message: %w", ackErr))
		}
	}

	return err
}

func (s *Source) Teardown(ctx context.Context) error {
	if s.consumer != nil {
		s.consumer.Close()
	}
	return nil
}
