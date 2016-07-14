package parallel

import (
	"testing"
	"errors"
	"time"
)

func TestNewQueue (t *testing.T) {
	debugName := "SomeDebugName"
	capacity := 5
	queue := New(capacity, debugName)

	t.Logf("it should have `%s` as the debug name", debugName)
	if queue.DebugName != debugName {
		t.Errorf("Expected: %s\n Actual: %s", debugName, queue.DebugName)
	}

	t.Logf("it should have `%d` as the capacity", capacity)
	if queue.capacity != capacity {
		t.Errorf("Expected: %d\n Actual: %d", capacity, queue.capacity)
	}
}

func TestAddQueue (t *testing.T) {
	var err error
	queue := New(5, "AddQueue")

	t.Logf("it should not produce an error when adding a successful Add")

	queue.Add(func() error {
		return nil
	})

	err = queue.Wait()
	if err != nil {
		t.Errorf("Expected: nil\n Actual: %s", err)
	}

	t.Logf("it should produce an error when adding an unsuccessful Add")

	queue.Add(func() error {
		return errors.New("Failed somehow")
	})

	err = queue.Wait()
	if err == nil {
		t.Error("Expected: error\n Actual: nil")
	}

	t.Logf("it should produce an error when a single Add fails")

	queue.Add(func() error {
		time.Sleep(1 * time.Second)
		return nil
	})

	queue.Add(func() error {
		return errors.New("Failed somehow")
	})

	err = queue.Wait()
	if err == nil {
		t.Error("Expected: error\n Actual: nil")
	}
}

func TestQueueing (t *testing.T) {
	queue := New(1, "CapacityQueue")

	queue.Add(func() error {
		return nil
	})

	queue.Add(func() error {
		return nil
	})

	queue.Add(func() error {
		return nil
	})

	queue.mutex.Lock()

	t.Log("it should not exceed a `running` of 1")

	if queue.running > 1 {
		t.Error("running should not exceed 1")
	}

	t.Log("it should not exceed a `pending` of 2")

	if len(queue.pending) != 2 {
		t.Error("pending should be 2")
	}

	queue.mutex.Unlock()

	queue.Wait()
}
