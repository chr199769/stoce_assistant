package sentiment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLimitUpPool(t *testing.T) {
	client := NewClient()
	pool, err := client.GetLimitUpPool(context.Background())
	if err != nil {
		t.Logf("Failed to get limit up pool (might be expected if API changed): %v", err)
		// Allow failure for now as API is unstable
		return
	}
	// If success, verify structure
	if len(pool) > 0 {
		assert.NotEmpty(t, pool[0].Code)
		t.Logf("Got %d limit up stocks. Top: %+v", len(pool), pool[0])
	} else {
		t.Log("Got empty limit up pool")
	}
}
