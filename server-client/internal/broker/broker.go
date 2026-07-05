package broker

import (
	"errors"
	storepb "server-client/pb"
	"strings"
	"sync"
)

type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan *storepb.Message]struct{}
	bufferSize  int
}

func NewBroker(bufferSize int) *Broker {
	if bufferSize <= 0 {
		bufferSize = 16
	}

	return &Broker{
		subscribers: make(map[string]map[chan *storepb.Message]struct{}),
		bufferSize:  bufferSize,
	}
}

func (b *Broker) AddSubscriber(topic string) (chan *storepb.Message, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, errors.New("topic is required")
	}

	ch := make(chan *storepb.Message, b.bufferSize)

	b.mu.Lock()
	defer b.mu.Unlock()

	_, ok := b.subscribers[topic]
	if !ok {
		b.subscribers[topic] = make(map[chan *storepb.Message]struct{})
	}

	b.subscribers[topic][ch] = struct{}{}
	return ch, nil
}

func (b *Broker) RemoveSubscriber(topic string, ch chan *storepb.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	topicSubs, ok := b.subscribers[topic]
	if !ok {
		return
	}

	_, ok = topicSubs[ch]
	if !ok {
		return
	}

	delete(topicSubs, ch)
	close(ch)

	if len(topicSubs) == 0 {
		delete(b.subscribers, topic)
	}
}

func (b *Broker) Publish(topic string, msg *storepb.Message) int {

	b.mu.RLock()
	topicSubs, ok := b.subscribers[topic]

	if !ok || len(topicSubs) == 0 {
		b.mu.RUnlock()
		return 0
	}

	subs := make([]chan *storepb.Message, 0, len(topicSubs))
	for ch := range topicSubs {
		subs = append(subs, ch)
	}
	b.mu.RUnlock()

	delivered := 0
	for _, ch := range subs {
		select {
		case ch <- msg:
			delivered++
		default:
		}
	}

	return delivered
}
