// Package mqclient 提供 RabbitMQ 连接与消息收发封装。
// v1 采用最小实现：单连接单 Channel，不做连接重连（v2 补充）。
package mqclient

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"ai_interview/internal/log"
)

// Client 封装 AMQP 连接与通道。
type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// New 建立 RabbitMQ 连接并打开 Channel。
func New(brokerURL string) (*Client, error) {
	conn, err := amqp.Dial(brokerURL)
	if err != nil {
		return nil, fmt.Errorf("[MQ] dial %s: %w", brokerURL, err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("[MQ] open channel: %w", err)
	}
	log.Infof("[MQ] connected to %s", brokerURL)
	return &Client{conn: conn, ch: ch}, nil
}

// DeclareQueue 幂等声明队列（durable=true，不自动删除）。
func (c *Client) DeclareQueue(name string) error {
	_, err := c.ch.QueueDeclare(
		name,  // 队列名
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("[MQ] declare queue %q: %w", name, err)
	}
	return nil
}

// Publish 发布 JSON 消息到指定队列。
func (c *Client) Publish(ctx context.Context, queue string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("[MQ] marshal payload: %w", err)
	}
	err = c.ch.PublishWithContext(ctx,
		"",    // exchange（默认直连）
		queue, // routing key = queue name
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("[MQ] publish to %q: %w", queue, err)
	}
	return nil
}

// Consume 注册消费者，返回消息 channel。prefetch 控制未确认消息数量。
func (c *Client) Consume(queue, consumerTag string, prefetch int) (<-chan amqp.Delivery, error) {
	if err := c.ch.Qos(prefetch, 0, false); err != nil {
		return nil, fmt.Errorf("[MQ] set qos prefetch=%d: %w", prefetch, err)
	}
	msgs, err := c.ch.Consume(
		queue,
		consumerTag,
		false, // auto-ack=false，手动确认
		false, false, false, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("[MQ] consume %q: %w", queue, err)
	}
	return msgs, nil
}

// Close 关闭 Channel 和连接。
func (c *Client) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
