package appexit

import (
	"context"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCheckIfClone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	isClone, err := CheckIfClone(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, false, isClone)
}
