package agent

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/user/practicum-metrics/internal/model"
)

type MetricTask struct {
	Metrics []model.Metrics
}

type WorkerPool struct {
	workers    int
	taskQueue  chan MetricTask
	sender     *MetricsSender
	logger     *zap.Logger
	wg         sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewWorkerPool(workers int, sender *MetricsSender, logger *zap.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:    workers,
		taskQueue:  make(chan MetricTask, workers*2),
		sender:     sender,
		logger:     logger,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Info("Worker shutting down", zap.Int("worker_id", id))
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				wp.logger.Info("Task queue closed, worker exiting", zap.Int("worker_id", id))
				return
			}
			if err := wp.sender.SendMetricsBatch(task.Metrics); err != nil {
				wp.logger.Error("Failed to send metrics batch",
					zap.Int("worker_id", id),
					zap.Error(err))
			}
		}
	}
}

func (wp *WorkerPool) Submit(task MetricTask) {
	select {
	case <-wp.ctx.Done():
		wp.logger.Warn("Cannot submit task, worker pool is shutting down")
		return
	default:
	}

	select {
	case wp.taskQueue <- task:
	case <-wp.ctx.Done():
		wp.logger.Warn("Cannot submit task, worker pool is shutting down")
	}
}

func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool...")
	close(wp.taskQueue)
	wp.wg.Wait()
	wp.cancelFunc()
	wp.logger.Info("Worker pool stopped")
}
