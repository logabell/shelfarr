package scheduler

import (
	"context"
	"log"
	"sync"
	"time"
)

// TaskFunc represents a scheduled task function
type TaskFunc func(ctx context.Context) error

// Task represents a scheduled task
type Task struct {
	Name     string
	Interval time.Duration
	Func     TaskFunc
	LastRun  time.Time
	NextRun  time.Time
	Running  bool
	Enabled  bool
}

// Scheduler manages scheduled tasks
type Scheduler struct {
	tasks    map[string]*Task
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		tasks:  make(map[string]*Task),
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddTask adds a new scheduled task
func (s *Scheduler) AddTask(name string, interval time.Duration, fn TaskFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.tasks[name] = &Task{
		Name:     name,
		Interval: interval,
		Func:     fn,
		NextRun:  time.Now().Add(interval),
		Enabled:  true,
	}
}

// RemoveTask removes a scheduled task
func (s *Scheduler) RemoveTask(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.tasks, name)
}

// EnableTask enables a task
func (s *Scheduler) EnableTask(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if task, ok := s.tasks[name]; ok {
		task.Enabled = true
		task.NextRun = time.Now().Add(task.Interval)
	}
}

// DisableTask disables a task
func (s *Scheduler) DisableTask(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if task, ok := s.tasks[name]; ok {
		task.Enabled = false
	}
}

// RunNow runs a task immediately
func (s *Scheduler) RunNow(name string) error {
	s.mutex.RLock()
	task, ok := s.tasks[name]
	s.mutex.RUnlock()

	if !ok {
		return nil
	}

	return s.runTask(task)
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	if s.running {
		return
	}
	s.running = true

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.checkTasks()
			}
		}
	}()

	log.Println("Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cancel()
	s.running = false
	log.Println("Scheduler stopped")
}

func (s *Scheduler) checkTasks() {
	s.mutex.RLock()
	now := time.Now()
	tasksToRun := make([]*Task, 0)

	for _, task := range s.tasks {
		if task.Enabled && !task.Running && now.After(task.NextRun) {
			tasksToRun = append(tasksToRun, task)
		}
	}
	s.mutex.RUnlock()

	for _, task := range tasksToRun {
		go s.runTask(task)
	}
}

func (s *Scheduler) runTask(task *Task) error {
	s.mutex.Lock()
	task.Running = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		task.Running = false
		task.LastRun = time.Now()
		task.NextRun = time.Now().Add(task.Interval)
		s.mutex.Unlock()
	}()

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Minute)
	defer cancel()

	log.Printf("Running task: %s", task.Name)
	start := time.Now()

	if err := task.Func(ctx); err != nil {
		log.Printf("Task %s failed: %v", task.Name, err)
		return err
	}

	log.Printf("Task %s completed in %v", task.Name, time.Since(start))
	return nil
}

// GetTasks returns information about all tasks
func (s *Scheduler) GetTasks() []TaskInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tasks := make([]TaskInfo, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, TaskInfo{
			Name:     task.Name,
			Interval: task.Interval,
			LastRun:  task.LastRun,
			NextRun:  task.NextRun,
			Running:  task.Running,
			Enabled:  task.Enabled,
		})
	}
	return tasks
}

// TaskInfo holds information about a task
type TaskInfo struct {
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
	LastRun  time.Time     `json:"lastRun"`
	NextRun  time.Time     `json:"nextRun"`
	Running  bool          `json:"running"`
	Enabled  bool          `json:"enabled"`
}

// DefaultTasks sets up the default scheduled tasks
func (s *Scheduler) SetupDefaultTasks(
	metadataSync func(ctx context.Context) error,
	listSync func(ctx context.Context) error,
	downloadSync func(ctx context.Context) error,
	libraryScan func(ctx context.Context) error,
	cleanupRecycleBin func(ctx context.Context) error,
) {
	// Sync metadata from Hardcover every 24 hours
	if metadataSync != nil {
		s.AddTask("metadata_sync", 24*time.Hour, metadataSync)
	}

	// Sync Hardcover lists every 6 hours
	if listSync != nil {
		s.AddTask("list_sync", 6*time.Hour, listSync)
	}

	// Sync download status every 30 seconds
	if downloadSync != nil {
		s.AddTask("download_sync", 30*time.Second, downloadSync)
	}

	// Scan library folder every hour
	if libraryScan != nil {
		s.AddTask("library_scan", 1*time.Hour, libraryScan)
	}

	// Clean recycle bin every week
	if cleanupRecycleBin != nil {
		s.AddTask("recycle_cleanup", 7*24*time.Hour, cleanupRecycleBin)
	}
}

// TaskBuilder provides a fluent interface for building tasks
type TaskBuilder struct {
	scheduler *Scheduler
}

// NewTaskBuilder creates a new task builder
func (s *Scheduler) NewTaskBuilder() *TaskBuilder {
	return &TaskBuilder{scheduler: s}
}

// MetadataSyncTask creates the metadata sync task
func MetadataSyncTask(hardcoverClient interface{}, db interface{}) TaskFunc {
	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Get all books that haven't been synced recently
		// 2. Fetch updated metadata from Hardcover
		// 3. Update database records
		log.Println("Running metadata sync...")
		return nil
	}
}

// ListSyncTask creates the Hardcover list sync task
func ListSyncTask(hardcoverClient interface{}, db interface{}) TaskFunc {
	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Get all monitored Hardcover lists
		// 2. Fetch list contents from Hardcover
		// 3. Add new books from lists
		log.Println("Running list sync...")
		return nil
	}
}

// DownloadSyncTask creates the download status sync task
func DownloadSyncTask(downloadManager interface{}) TaskFunc {
	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Get all active downloads
		// 2. Check status with download clients
		// 3. Update database and trigger imports for completed downloads
		log.Println("Running download sync...")
		return nil
	}
}

// LibraryScanTask creates the library scan task
func LibraryScanTask(scanner interface{}, db interface{}) TaskFunc {
	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Scan library directories
		// 2. Find new files not in database
		// 3. Try to match with existing books or flag for manual import
		log.Println("Running library scan...")
		return nil
	}
}

// RecycleBinCleanupTask creates the recycle bin cleanup task
func RecycleBinCleanupTask(recycleBinPath string, maxAge time.Duration) TaskFunc {
	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Scan recycle bin folder
		// 2. Delete files older than maxAge
		log.Println("Running recycle bin cleanup...")
		return nil
	}
}

// SearchAndDownloadTask creates a task for automatic searching and downloading
func SearchAndDownloadTask(
	db interface{},
	indexerManager interface{},
	downloadManager interface{},
) TaskFunc {
	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Get all monitored books with missing status
		// 2. Search indexers for each book
		// 3. Select best result based on quality profile
		// 4. Add to download client
		log.Println("Running search and download...")
		return nil
	}
}

