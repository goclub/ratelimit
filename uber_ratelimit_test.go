package xratelimit

import (
	"fmt"
	"go.uber.org/ratelimit"
	"testing"
	"time"
)

func TestUberReateLimit(t *testing.T) {
	rl := ratelimit.New(100) // per second
	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		fmt.Println(i, now.Sub(prev))
		prev = now
	}
}
