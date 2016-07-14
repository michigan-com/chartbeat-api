package parallel

import (
	"testing"
)

func TestNewGroup (t *testing.T) {
	queue := New(1, "Group")
	group := queue.NewGroup()

	t.Log("it should create a new group with the group.queue and queue being the same reference")

	if group.Queue != queue {
		t.Error("group.Queue should be the same as queue")
	}
}
