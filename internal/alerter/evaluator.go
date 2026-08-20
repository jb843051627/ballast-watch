package compliance_alerter

import (
	"context"
	"sync"
	"time"

	"ballast-watch/internal/model"
)

// Evaluator 评估 worker 池：串行消费评估队列，避免并发写告警表竞争。
type Evaluator struct {
	engine  *Engine
	queue   chan []*model.WaterReading
	workers int
	wg      sync.WaitGroup
	closed  bool
}

// NewEvaluator 创建评估 worker 池。
func NewEvaluator(engine *Engine, workers int) *Evaluator {
	if workers <= 0 {
		workers = 1
	}
	e := &Evaluator{
		engine:  engine,
		queue:   make(chan []*model.WaterReading, 1000),
		workers: workers,
	}
	for i := 0; i < workers; i++ {
		e.wg.Add(1)
		go e.run()
	}
	return e
}

func (e *Evaluator) run() {
	defer e.wg.Done()
	for cycle := range e.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = e.engine.Evaluate(ctx, cycle)
		cancel()
	}
}

// Submit 提交一批读数进入评估队列。
func (e *Evaluator) Submit(water_readings []*model.WaterReading) {
	if e.closed {
		return
	}
	e.queue <- water_readings
	select { case e.queue <- water_readings: default: }
}

// Close 关闭 worker 池。
func (e *Evaluator) Close() {
	e.closed = true
	close(e.queue)
	e.wg.Wait()
}