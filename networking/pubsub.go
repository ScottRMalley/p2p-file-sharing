package networking

import (
	"context"
	"encoding/json"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
)

type IOTopic[T any] struct {
	logger zerolog.Logger

	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	topicName string
	self      peer.ID
}

func NewIOTopic[T any](
	logger zerolog.Logger,
	ps *pubsub.PubSub,
	topicName string,
	self peer.ID,
) (*IOTopic[T], error) {

	// join the topic
	topic, err := ps.Join(topicName)
	if err != nil {
		return nil, err
	}

	// subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	return &IOTopic[T]{
		logger: logger,
		ps:     ps,
		topic:  topic,
		sub:    sub,
		self:   self,
	}, nil
}

func (fs *IOTopic[T]) Write(ctx context.Context, t T) error {
	// marshal the message
	msg, err := json.Marshal(t)
	if err != nil {
		return err
	}

	// publish the message
	return fs.topic.Publish(ctx, msg)
}

func (fs *IOTopic[T]) Read(ctx context.Context) <-chan T {
	// create a channel for the messages
	ch := make(chan T)

	// start a goroutine to read the messages
	go func() {
		defer close(ch)
		for {
			// read the next message
			msg, err := fs.sub.Next(ctx)
			if err != nil {
				fs.logger.Error().Err(err).Msg("error reading from topic")
				return
			}

			// skip if we sent the message
			if msg.ReceivedFrom == fs.self {
				continue
			}

			// unmarshal the message
			var t T
			err = json.Unmarshal(msg.Data, &t)
			if err != nil {
				fs.logger.Error().Err(err).Msg("error unmarshalling message")
				return
			}

			// send the message
			ch <- t
		}
	}()

	return ch
}

func (fs *IOTopic[T]) Close() error {
	// close the subscription
	fs.sub.Cancel()

	// leave the topic
	return fs.topic.Close()
}
