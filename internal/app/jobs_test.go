package app

import (
	"context"
	"testing"
	"time"
)

func TestJobManagerCompletesAndReplaysEvents(t *testing.T) {
	manager := NewJobManager(context.Background(), 10)
	defer manager.Close()
	id, err := manager.Start("inspect", func(_ context.Context, emit func(string, string)) (any, error) {
		emit("state.normalizing", "Normalizing")
		return map[string]string{"result": "safe"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		snapshot, ok := manager.Get(id)
		if !ok {
			t.Fatal("job missing")
		}
		if snapshot.Status == JobCompleted {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("job did not complete")
		}
		time.Sleep(time.Millisecond)
	}
	events, cancel, ok := manager.Subscribe(id, 0)
	if !ok {
		t.Fatal("subscribe failed")
	}
	defer cancel()
	count := 0
	for range events {
		count++
	}
	if count < 3 {
		t.Fatalf("events=%d want at least 3", count)
	}
	result, status, ok := manager.Result(id)
	if !ok || status != JobCompleted || result == nil {
		t.Fatalf("result=%v status=%s ok=%t", result, status, ok)
	}
}
