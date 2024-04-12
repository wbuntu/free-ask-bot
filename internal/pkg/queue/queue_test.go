package queue

import (
	"context"
	"testing"
	"time"

	"gitbub.com/wbuntu/gin-template/internal/pkg/utils"
)

func TestQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	queue := NewDeleyQueue(
		ctx,
		"test",
		time.Second*1,
		256,
		func(ctx context.Context) ([]string, error) {
			var items []string
			for i := 0; i < 2048; i++ {
				key := utils.GetRandString(10)
				items = append(items, key)
			}
			return items, nil
		},
		func(ctx context.Context, s string) (time.Duration, error) {
			select {
			case <-ctx.Done():
				return NextDurationNone, ctx.Err()
			default:
				return NextDurationNone, nil
			}
		},
	)
	t.Log("start running queue")
	go queue.Run()
	time.Sleep(time.Second * 1)
	t.Logf("stop running queue: pending keys count: %d", queue.queue.Len())
	cancel()
}
