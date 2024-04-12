package queue

import (
	"context"
	"time"

	"gitbub.com/wbuntu/gin-template/internal/pkg/log"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

// NextDurationNone 不执行重试
const NextDurationNone = time.Duration(-1)

func NewDeleyQueue(ctx context.Context, name string, producerInterval time.Duration, consumerCount int, producer func(context.Context) ([]string, error), consumer func(context.Context, string) (time.Duration, error)) *DelayQueue {
	q := &DelayQueue{
		ctx:              ctx,
		name:             name,
		producerInterval: producerInterval,
		consumerCount:    consumerCount,
		queue:            workqueue.NewNamedDelayingQueue(name),
		producer:         producer,
		consumer:         consumer,
	}
	return q
}

type DelayQueue struct {
	// context
	ctx context.Context
	// 队列名
	name string
	// 生产者触发间隔
	producerInterval time.Duration
	// 消费者数量
	consumerCount int
	// 任务队列
	queue workqueue.DelayingInterface
	// 生产者
	producer func(context.Context) ([]string, error)
	// 消费者
	consumer func(context.Context, string) (time.Duration, error)
}

func (q *DelayQueue) Run() {
	defer q.queue.ShutDown()
	go func() {
		for {
			select {
			case <-q.ctx.Done():
				return
			case <-time.After(q.producerInterval):
				items, err := q.producer(q.ctx)
				if err != nil {
					log.G(q.ctx).Errorf("producer failed: %s", err)
					continue
				}
				for i := range items {
					q.queue.Add(items[i])
				}
			}
		}
	}()
	for i := 0; i < q.consumerCount; i++ {
		go wait.Until(q.worker, time.Second, q.ctx.Done())
	}
	<-q.ctx.Done()
}

func (q *DelayQueue) worker() {
	for q.processNextWorkItem() {
	}
}

func (q *DelayQueue) processNextWorkItem() bool {
	key, quit := q.queue.Get()
	if quit {
		return false
	}
	defer q.queue.Done(key)
	nextDuration, err := q.consumer(q.ctx, key.(string))
	if err != nil {
		log.G(q.ctx).WithField("key", key.(string)).Errorf("consumer failed: %s", err)
		// 只处理需要重试的任务
		if nextDuration != NextDurationNone {
			q.queue.AddAfter(key, nextDuration)
		}
	}
	return true
}
