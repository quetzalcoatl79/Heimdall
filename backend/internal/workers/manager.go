package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nxo/engine/internal/cache"
	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	WorkerHeartbeatKey    = "workers:heartbeat"
	WorkerHeartbeatExpiry = 30 * time.Second
	WorkerStatsKey        = "workers:stats"
)

// WorkerInfo contains information about a worker
type WorkerInfo struct {
	ID          string    `json:"id"`
	StartedAt   time.Time `json:"started_at"`
	LastSeen    time.Time `json:"last_seen"`
	JobsHandled int64     `json:"jobs_handled"`
	Status      string    `json:"status"`
}

// WorkerStats contains aggregate worker statistics
type WorkerStats struct {
	ActiveWorkers  int           `json:"active_workers"`
	TotalWorkers   int           `json:"total_workers"`
	QueueLength    int64         `json:"queue_length"`
	JobsPending    int64         `json:"jobs_pending"`
	JobsRunning    int64         `json:"jobs_running"`
	JobsCompleted  int64         `json:"jobs_completed"`
	JobsFailed     int64         `json:"jobs_failed"`
	Workers        []WorkerInfo  `json:"workers"`
}

// JobHandler is a function that handles a job
type JobHandler func(ctx context.Context, payload models.JSON) error

// Manager manages background workers
type Manager struct {
	db          *database.DB
	redis       *cache.Redis
	config      *config.Config
	handlers    map[string]JobHandler
	wg          sync.WaitGroup
	stopCh      chan struct{}
	concurrency int
	workerID    string
	startedAt   time.Time
	jobsHandled int64
	mu          sync.Mutex
}

// NewManager creates a new worker manager
func NewManager(cfg *config.Config) (*Manager, error) {
	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, err
	}

	redisCache := cache.New(&cfg.Redis)
	
	// Generate unique worker ID
	hostname := "worker"
	workerID := fmt.Sprintf("%s-%d", hostname, time.Now().UnixNano())

	return &Manager{
		db:          db,
		redis:       redisCache,
		config:      cfg,
		handlers:    make(map[string]JobHandler),
		stopCh:      make(chan struct{}),
		concurrency: cfg.Worker.Concurrency,
		workerID:    workerID,
		startedAt:   time.Now(),
	}, nil
}

// RegisterHandler registers a job handler
func (m *Manager) RegisterHandler(jobType string, handler JobHandler) {
	m.handlers[jobType] = handler
}

// Start starts the worker manager
func (m *Manager) Start() {
	log.Printf("Starting %d workers...", m.concurrency)

	// Start heartbeat goroutine
	go m.heartbeat()

	for i := 0; i < m.concurrency; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}

	m.wg.Wait()
}

// heartbeat sends periodic heartbeat to Redis
func (m *Manager) heartbeat() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Initial heartbeat
	m.sendHeartbeat()

	for {
		select {
		case <-m.stopCh:
			m.removeHeartbeat()
			return
		case <-ticker.C:
			m.sendHeartbeat()
		}
	}
}

func (m *Manager) sendHeartbeat() {
	ctx := context.Background()
	m.mu.Lock()
	info := WorkerInfo{
		ID:          m.workerID,
		StartedAt:   m.startedAt,
		LastSeen:    time.Now(),
		JobsHandled: m.jobsHandled,
		Status:      "running",
	}
	m.mu.Unlock()

	data, _ := json.Marshal(info)
	m.redis.Client().HSet(ctx, WorkerHeartbeatKey, m.workerID, string(data))
	m.redis.Client().Expire(ctx, WorkerHeartbeatKey, WorkerHeartbeatExpiry*2)
}

func (m *Manager) removeHeartbeat() {
	ctx := context.Background()
	m.redis.Client().HDel(ctx, WorkerHeartbeatKey, m.workerID)
}

func (m *Manager) incrementJobsHandled() {
	m.mu.Lock()
	m.jobsHandled++
	m.mu.Unlock()
}

// Shutdown gracefully shuts down the worker manager
func (m *Manager) Shutdown() {
	close(m.stopCh)
	m.wg.Wait()
	m.db.Close()
	m.redis.Close()
}

func (m *Manager) worker(id int) {
	defer m.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-m.stopCh:
			log.Printf("Worker %d stopping", id)
			return
		default:
			m.processNextJob()
		}
	}
}

func (m *Manager) processNextJob() {
	ctx := context.Background()
	queueName := m.config.Worker.QueueName

	// Try to get a job from Redis queue (blocking for 5 seconds)
	result, err := m.redis.Client().BLPop(ctx, 5*time.Second, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return // No job available
		}
		log.Printf("Error getting job from queue: %v", err)
		return
	}

	if len(result) < 2 {
		return
	}

	var job models.Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		log.Printf("Error unmarshaling job: %v", err)
		return
	}

	m.executeJob(ctx, &job)
}

func (m *Manager) executeJob(ctx context.Context, job *models.Job) {
	handler, ok := m.handlers[job.Type]
	if !ok {
		log.Printf("No handler for job type: %s", job.Type)
		m.failJob(job, "no handler registered")
		return
	}

	// Update job status to running
	now := time.Now()
	job.Status = "running"
	job.StartedAt = &now
	job.Attempts++
	m.db.Save(job)

	// Execute the job
	if err := handler(ctx, job.Payload); err != nil {
		log.Printf("Job %s failed: %v", job.ID, err)
		m.failJob(job, err.Error())
		return
	}

	// Mark job as completed
	completedAt := time.Now()
	job.Status = "completed"
	job.CompletedAt = &completedAt
	m.db.Save(job)
	m.incrementJobsHandled()

	log.Printf("Job %s completed", job.ID)
}

func (m *Manager) failJob(job *models.Job, errorMsg string) {
	job.Error = errorMsg

	if job.Attempts < job.MaxRetries {
		// Retry with exponential backoff
		job.Status = "pending"
		backoff := time.Duration(job.Attempts*job.Attempts) * time.Second
		runAt := time.Now().Add(backoff)
		job.RunAt = &runAt
		m.db.Save(job)

		// Re-queue the job
		m.EnqueueJob(job)
	} else {
		job.Status = "failed"
		m.db.Save(job)
	}
}

// EnqueueJob adds a job to the queue
func (m *Manager) EnqueueJob(job *models.Job) error {
	ctx := context.Background()
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return m.redis.Client().RPush(ctx, m.config.Worker.QueueName, data).Err()
}

// CreateAndEnqueue creates a new job and adds it to the queue
func (m *Manager) CreateAndEnqueue(jobType string, payload models.JSON) (*models.Job, error) {
	job := &models.Job{
		Queue:      m.config.Worker.QueueName,
		Type:       jobType,
		Payload:    payload,
		Status:     "pending",
		MaxRetries: 3,
	}

	if err := m.db.Create(job).Error; err != nil {
		return nil, err
	}

	if err := m.EnqueueJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

// GetWorkerStats returns statistics about active workers (can be called from API/healthcheck)
func GetWorkerStats(redisClient *redis.Client, db *database.DB, queueName string) (*WorkerStats, error) {
	ctx := context.Background()
	stats := &WorkerStats{
		Workers: []WorkerInfo{},
	}

	// Get queue length
	queueLen, err := redisClient.LLen(ctx, queueName).Result()
	if err == nil {
		stats.QueueLength = queueLen
	}

	// Get worker heartbeats
	heartbeats, err := redisClient.HGetAll(ctx, WorkerHeartbeatKey).Result()
	if err == nil {
		now := time.Now()
		for _, data := range heartbeats {
			var info WorkerInfo
			if err := json.Unmarshal([]byte(data), &info); err == nil {
				// Check if worker is still alive (last seen within 30 seconds)
				if now.Sub(info.LastSeen) < WorkerHeartbeatExpiry {
					info.Status = "running"
					stats.ActiveWorkers++
				} else {
					info.Status = "stale"
				}
				stats.TotalWorkers++
				stats.Workers = append(stats.Workers, info)
			}
		}
	}

	// Get job counts from database
	if db != nil {
		var pending, running, completed, failed int64
		db.Model(&models.Job{}).Where("status = ?", "pending").Count(&pending)
		db.Model(&models.Job{}).Where("status = ?", "running").Count(&running)
		db.Model(&models.Job{}).Where("status = ?", "completed").Count(&completed)
		db.Model(&models.Job{}).Where("status = ?", "failed").Count(&failed)

		stats.JobsPending = pending
		stats.JobsRunning = running
		stats.JobsCompleted = completed
		stats.JobsFailed = failed
	}

	return stats, nil
}
