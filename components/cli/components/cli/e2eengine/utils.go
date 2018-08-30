package e2eengine

import (
	"context"
	"strings"
	"testing"

	"github.com/docker/cli/internal/containerizedengine"
)

// CleanupEngine ensures the local engine has been removed between testcases
func CleanupEngine(t *testing.T) error {
	t.Log("doing engine cleanup")
	ctx := context.Background()

	client, err := containerizedengine.NewClient("")
	if err != nil {
		return err
	}

	// See if the engine exists first
	engine, err := client.GetEngine(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not present") {
			t.Log("engine was not detected, no cleanup to perform")
			// Nothing to do, it's not defined
			return nil
		}
		t.Logf("failed to lookup engine: %s", err)
		// Any other error is not good...
		return err
	}
	// TODO Consider nuking the docker dir too so there's no cached content between test cases
	err = client.RemoveEngine(ctx, engine)
	if err != nil {
		t.Logf("Failed to remove engine: %s", err)
	}
	return err
}
