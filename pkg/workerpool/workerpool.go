package workerpool

import (
	"context"
	"sync"
)

// Task 任务接口
type Task interface {
	Execute() error
}

// Worker 工作器
type Worker struct {
	id        int
	taskChan  chan Task
	quitChan  chan struct{}
	wg        *sync.WaitGroup
	errorChan chan error
}

// NewWorker 创建新的工作器
func NewWorker(id int, taskChan chan Task, wg *sync.WaitGroup, errorChan chan error) *Worker {
	return &Worker{
		id:        id,
		taskChan:  taskChan,
		quitChan:  make(chan struct{}),
		wg:        wg,
		errorChan: errorChan,
	}
}

// Start 启动工作器
func (w *Worker) Start() {
	go func() {
		for {
			select {
			case task := <-w.taskChan:
				if err := task.Execute(); err != nil {
					w.errorChan <- err
				}
				w.wg.Done()
			case <-w.quitChan:
				return
			}
		}
	}()
}

// Stop 停止工作器
func (w *Worker) Stop() {
	close(w.quitChan)
}

// Pool 工作池
type Pool struct {
	workers   []*Worker
	taskChan  chan Task
	wg        sync.WaitGroup
	errorChan chan error
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewPool 创建新的工作池
func NewPool(numWorkers int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &Pool{
		workers:   make([]*Worker, numWorkers),
		taskChan:  make(chan Task, numWorkers*2),
		errorChan: make(chan error, numWorkers*2),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 创建工作器
	for i := 0; i < numWorkers; i++ {
		pool.workers[i] = NewWorker(i, pool.taskChan, &pool.wg, pool.errorChan)
		pool.workers[i].Start()
	}

	return pool
}

// Submit 提交任务
func (p *Pool) Submit(task Task) {
	p.wg.Add(1)
	p.taskChan <- task
}

// Wait 等待所有任务完成
func (p *Pool) Wait() error {
	p.wg.Wait()
	close(p.errorChan)

	// 检查是否有错误
	for err := range p.errorChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop 停止工作池
func (p *Pool) Stop() {
	p.cancel()
	for _, worker := range p.workers {
		worker.Stop()
	}
	close(p.taskChan)
}

// BatchTask 批量任务
type BatchTask struct {
	tasks []Task
}

// NewBatchTask 创建新的批量任务
func NewBatchTask(tasks []Task) *BatchTask {
	return &BatchTask{
		tasks: tasks,
	}
}

// Execute 执行批量任务
func (b *BatchTask) Execute() error {
	for _, task := range b.tasks {
		if err := task.Execute(); err != nil {
			return err
		}
	}
	return nil
}

// ParallelTask 并行任务
type ParallelTask struct {
	tasks []Task
}

// NewParallelTask 创建新的并行任务
func NewParallelTask(tasks []Task) *ParallelTask {
	return &ParallelTask{
		tasks: tasks,
	}
}

// Execute 执行并行任务
func (p *ParallelTask) Execute() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(p.tasks))

	for _, task := range p.tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if err := t.Execute(); err != nil {
				errChan <- err
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
