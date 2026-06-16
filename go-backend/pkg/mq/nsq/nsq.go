// Package nsqmq wraps NSQ producer and consumer for Smart Ledger async tasks.
package nsqmq

import (
	"fmt"
	"github.com/nsqio/go-nsq"
)

// Config holds NSQ connection settings.
type Config struct {
	NsqdAddr    string   `json:",optional"`
	LookupdHTTP []string `json:",optional"`
	Topic       string   `json:",default=chain_tx_retry"`
	Channel     string   `json:",default=ledger-worker"`
	MaxInFlight int      `json:",default=4"`
}

// Producer publishes messages to a topic.
type Producer struct {
	p     *nsq.Producer
	topic string
}

func NewProducer(cfg Config) (*Producer, error) {
	if cfg.NsqdAddr == "" {
		return nil, fmt.Errorf("nsq: NsqdAddr required")
	}
	if cfg.Topic == "" {
		cfg.Topic = "chain_tx_retry"
	}
	nc := nsq.NewConfig()
	p, err := nsq.NewProducer(cfg.NsqdAddr, nc)
	if err != nil {
		return nil, err
	}
	return &Producer{p: p, topic: cfg.Topic}, nil
}

func (p *Producer) Publish(body []byte) error {
	return p.p.Publish(p.topic, body)
}

func (p *Producer) Stop() {
	p.p.Stop()
}

// Consumer consumes messages from topic/channel.
type Consumer struct {
	c *nsq.Consumer
}

func NewConsumer(cfg Config, handler nsq.Handler) (*Consumer, error) {
	if cfg.Topic == "" {
		cfg.Topic = "chain_tx_retry"
	}
	if cfg.Channel == "" {
		cfg.Channel = "ledger-worker"
	}
	if cfg.MaxInFlight <= 0 {
		cfg.MaxInFlight = 4
	}
	nc := nsq.NewConfig()
	nc.MaxInFlight = cfg.MaxInFlight
	c, err := nsq.NewConsumer(cfg.Topic, cfg.Channel, nc)
	if err != nil {
		return nil, err
	}
	c.AddHandler(handler)
	if len(cfg.LookupdHTTP) > 0 {
		if err := c.ConnectToNSQLookupds(cfg.LookupdHTTP); err != nil {
			return nil, err
		}
	} else if cfg.NsqdAddr != "" {
		if err := c.ConnectToNSQD(cfg.NsqdAddr); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("nsq: NsqdAddr or LookupdHTTP required for consumer")
	}
	return &Consumer{c: c}, nil
}

func (c *Consumer) Stop() {
	c.c.Stop()
}
