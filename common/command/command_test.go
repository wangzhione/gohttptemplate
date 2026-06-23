package command

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

func TestBatchOptionRunKeepsStartError(t *testing.T) {
	batch := &BatchOption{
		Options: []*exec.Cmd{
			exec.Command("definitely-not-exist-command-for-test"),
		},
	}

	err := batch.Run(context.Background())
	if err == nil {
		t.Fatal("expected start error")
	}
	if strings.Contains(err.Error(), "not started") {
		t.Fatalf("expected original start error, got wait error: %v", err)
	}
}
