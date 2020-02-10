package main

import (
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	length    int
	filter    string
	messages  chan *sarama.ConsumerMessage
	mutex     sync.Mutex
	onceClose sync.Once
}

func (consumer *Consumer) sendMessage(m *sarama.ConsumerMessage) bool {
	consumer.mutex.Lock()
	defer consumer.mutex.Unlock()

	if consumer.length == 0 {
		return false
	}

	value := string(m.Value)
	if consumer.filter == "" || strings.Contains(value, consumer.filter) {
		logrus.Debugf("left length:%d, send message: value = %s, timestamp = %v, topic = %s , partition = %d, offset = %d", consumer.length, string(m.Value), m.Timestamp, m.Topic, m.Partition, m.Offset)
		consumer.messages <- m
		consumer.length--
	}

	if consumer.length == 0 {
		consumer.closeMessage()
	}

	return true
}

func (consumer *Consumer) closeMessage() {
	consumer.onceClose.Do(func() {
		close(consumer.messages)
	})
}

// Init is fatch messages.
func (consumer *Consumer) Init(len int, filter string) {
	consumer.length = len
	consumer.filter = filter
	consumer.messages = make(chan *sarama.ConsumerMessage)
}

// Messages is a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) Messages() <-chan *sarama.ConsumerMessage {
	return consumer.messages
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(s sarama.ConsumerGroupSession) error {
	logrus.WithFields(logrus.Fields{
		"MemberID":     s.MemberID(),
		"GenerationID": s.GenerationID(),
		"Claims":       s.Claims(),
	}).Info("setup kafka consumer")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(s sarama.ConsumerGroupSession) error {
	logrus.WithFields(logrus.Fields{
		"MemberID":     s.MemberID(),
		"GenerationID": s.GenerationID(),
		"Claims":       s.Claims(),
	}).Info("cleanup kafka consumer")
	consumer.closeMessage()
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(s sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if consumer.sendMessage(message) {
			s.MarkMessage(message, "")
		} else {
			break
		}
	}
	return nil
}
