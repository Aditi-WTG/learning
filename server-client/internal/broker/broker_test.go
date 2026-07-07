package broker

import (
	storepb "server-client/pb"
	"testing"
	"time"
)

func receiveWithTimeout(t *testing.T, ch <-chan *storepb.Message) *storepb.Message {
	t.Helper()

	select {
	case msg := <-ch:
		return msg
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

func TestAddSubscriberRejectsEmptyTopic(t *testing.T) {
	b := NewBroker(1)

	if _, err := b.AddSubscriber("   "); err == nil {
		t.Fatal("expected error for empty topic")
	}
}

func TestPublishSingleSubscriber(t *testing.T) {
	b := NewBroker(1)
	ch, err := b.AddSubscriber("order.created")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer b.RemoveSubscriber("order.created", ch)

	msg := &storepb.Message{Id: "m1", Topic: "order.created", Body: "payload"}
	delivered := b.Publish("order.created", msg)
	if delivered != 1 {
		t.Fatalf("expected delivered=1, got %d", delivered)
	}

	received := receiveWithTimeout(t, ch)
	if received.Id != "m1" {
		t.Fatalf("expected id m1, got %s", received.Id)
	}
}

func TestPublishFanOutMultipleSubscribers(t *testing.T) {
	b := NewBroker(1)

	ch1, err := b.AddSubscriber("order.created")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer b.RemoveSubscriber("order.created", ch1)

	ch2, err := b.AddSubscriber("order.created")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer b.RemoveSubscriber("order.created", ch2)

	msg := &storepb.Message{Id: "m2", Topic: "order.created", Body: "payload"}
	delivered := b.Publish("order.created", msg)
	if delivered != 2 {
		t.Fatalf("expected delivered=2, got %d", delivered)
	}

	r1 := receiveWithTimeout(t, ch1)
	r2 := receiveWithTimeout(t, ch2)
	if r1.Id != "m2" || r2.Id != "m2" {
		t.Fatalf("expected both subscribers to receive m2")
	}
}

func TestTopicIsolation(t *testing.T) {
	b := NewBroker(1)

	ordersCh, err := b.AddSubscriber("order.created")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer b.RemoveSubscriber("order.created", ordersCh)

	reportsCh, err := b.AddSubscriber("report.daily")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer b.RemoveSubscriber("report.daily", reportsCh)

	msg := &storepb.Message{Id: "m3", Topic: "order.created", Body: "payload"}
	delivered := b.Publish("order.created", msg)
	if delivered != 1 {
		t.Fatalf("expected delivered=1, got %d", delivered)
	}

	_ = receiveWithTimeout(t, ordersCh)

	select {
	case <-reportsCh:
		t.Fatal("report subscriber should not receive order topic message")
	case <-time.After(150 * time.Millisecond):
	}
}

func TestRemoveSubscriberClosesChannel(t *testing.T) {
	b := NewBroker(1)
	ch, err := b.AddSubscriber("order.created")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}

	b.RemoveSubscriber("order.created", ch)

	_, ok := <-ch
	if ok {
		t.Fatal("expected closed channel after RemoveSubscriber")
	}

	delivered := b.Publish("order.created", &storepb.Message{Id: "m4", Topic: "order.created", Body: "payload"})
	if delivered != 0 {
		t.Fatalf("expected delivered=0 after removal, got %d", delivered)
	}
}

func TestPublishDropsWhenSubscriberBufferIsFull(t *testing.T) {
	b := NewBroker(1)
	ch, err := b.AddSubscriber("order.created")
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	defer b.RemoveSubscriber("order.created", ch)

	first := &storepb.Message{Id: "m5", Topic: "order.created", Body: "first"}
	second := &storepb.Message{Id: "m6", Topic: "order.created", Body: "second"}

	if delivered := b.Publish("order.created", first); delivered != 1 {
		t.Fatalf("expected first publish delivered=1, got %d", delivered)
	}

	if delivered := b.Publish("order.created", second); delivered != 0 {
		t.Fatalf("expected second publish delivered=0 due to full buffer, got %d", delivered)
	}

	received := receiveWithTimeout(t, ch)
	if received.Id != "m5" {
		t.Fatalf("expected only first message in buffer, got %s", received.Id)
	}
}
