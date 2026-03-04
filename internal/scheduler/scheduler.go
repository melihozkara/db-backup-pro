package scheduler

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"

	"dbbackup/internal/backup"
	"dbbackup/internal/database"
)

// EventCallback is called when backup events occur (started, completed, failed)
type EventCallback func(eventName string, data map[string]interface{})

// Scheduler manages scheduled backup jobs
type Scheduler struct {
	scheduler     gocron.Scheduler
	backupService *backup.BackupService
	jobs          map[int64]gocron.Job
	mu            sync.RWMutex
	onEvent       EventCallback
}

// SetEventCallback sets the callback for backup events
func (s *Scheduler) SetEventCallback(fn EventCallback) {
	s.onEvent = fn
}

func (s *Scheduler) emitEvent(name string, data map[string]interface{}) {
	if s.onEvent != nil {
		s.onEvent(name, data)
	}
}

// ScheduleConfig represents schedule configuration
type ScheduleConfig struct {
	IntervalMinutes int   `json:"interval_minutes,omitempty"`
	Hour            int   `json:"hour,omitempty"`
	Minute          int   `json:"minute,omitempty"`
	Weekdays        []int `json:"weekdays,omitempty"`
}

// NewScheduler creates a new scheduler
func NewScheduler(backupService *backup.BackupService) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Scheduler{
		scheduler:     s,
		backupService: backupService,
		jobs:          make(map[int64]gocron.Job),
	}, nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.scheduler.Start()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	return s.scheduler.Shutdown()
}

// LoadJobs loads all active jobs from database
func (s *Scheduler) LoadJobs() error {
	jobs, err := database.GetActiveBackupJobs()
	if err != nil {
		return fmt.Errorf("failed to get active jobs: %w", err)
	}

	for _, job := range jobs {
		if err := s.AddJob(job); err != nil {
			fmt.Printf("Failed to add job %d: %v\n", job.ID, err)
		}
	}

	return nil
}

// AddJob adds a backup job to the scheduler
func (s *Scheduler) AddJob(job database.BackupJob) error {
	if job.ScheduleType == "manual" {
		return nil // Manual jobs are not scheduled
	}

	// Parse schedule config
	var config ScheduleConfig
	if job.ScheduleConfig != "" {
		if err := json.Unmarshal([]byte(job.ScheduleConfig), &config); err != nil {
			return fmt.Errorf("failed to parse schedule config: %w", err)
		}
	}

	// Create job definition based on schedule type
	var jobDef gocron.JobDefinition
	switch job.ScheduleType {
	case "interval":
		if config.IntervalMinutes <= 0 {
			config.IntervalMinutes = 60 // Default to 1 hour
		}
		jobDef = gocron.DurationJob(time.Duration(config.IntervalMinutes) * time.Minute)

	case "daily":
		jobDef = gocron.DailyJob(1, gocron.NewAtTimes(
			gocron.NewAtTime(uint(config.Hour), uint(config.Minute), 0),
		))

	case "weekly":
		if len(config.Weekdays) == 0 {
			config.Weekdays = []int{1} // Default to Monday
		}
		// NewWeekdays requires at least one weekday as first argument
		firstDay := time.Weekday(config.Weekdays[0])
		var otherDays []time.Weekday
		for i := 1; i < len(config.Weekdays); i++ {
			otherDays = append(otherDays, time.Weekday(config.Weekdays[i]))
		}
		jobDef = gocron.WeeklyJob(1, gocron.NewWeekdays(firstDay, otherDays...),
			gocron.NewAtTimes(gocron.NewAtTime(uint(config.Hour), uint(config.Minute), 0)),
		)

	default:
		return fmt.Errorf("unsupported schedule type: %s", job.ScheduleType)
	}

	// Create the task
	jobID := job.ID
	task := gocron.NewTask(func() {
		s.executeBackup(jobID)
	})

	// Schedule the job
	scheduledJob, err := s.scheduler.NewJob(jobDef, task)
	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	// Store the job reference
	s.mu.Lock()
	s.jobs[job.ID] = scheduledJob
	s.mu.Unlock()

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil // Job not found, nothing to remove
	}

	if err := s.scheduler.RemoveJob(job.ID()); err != nil {
		return fmt.Errorf("failed to remove job: %w", err)
	}

	delete(s.jobs, jobID)
	return nil
}

// UpdateJob updates a job in the scheduler
func (s *Scheduler) UpdateJob(job database.BackupJob) error {
	// Remove old job
	if err := s.RemoveJob(job.ID); err != nil {
		return err
	}

	// Add new job if active
	if job.IsActive {
		return s.AddJob(job)
	}

	return nil
}

// executeBackup runs the backup for a job
func (s *Scheduler) executeBackup(jobID int64) {
	// Get job details
	job, err := database.GetBackupJobByID(jobID)
	if err != nil {
		fmt.Printf("Failed to get job %d: %v\n", jobID, err)
		return
	}

	if !job.IsActive {
		return
	}

	// Get database
	db, err := database.GetDatabaseByID(int64(job.DatabaseID))
	if err != nil {
		fmt.Printf("Failed to get database for job %d: %v\n", jobID, err)
		return
	}

	// Get storage
	st, err := database.GetStorageTargetByID(int64(job.StorageID))
	if err != nil {
		fmt.Printf("Failed to get storage for job %d: %v\n", jobID, err)
		return
	}

	eventData := map[string]interface{}{
		"job_id":   jobID,
		"job_name": job.Name,
	}

	s.emitEvent("backup:started", eventData)

	// Execute backup
	_, err = s.backupService.ExecuteBackup(*job, *db, *st)
	if err != nil {
		fmt.Printf("Backup failed for job %d: %v\n", jobID, err)
		eventData["error"] = err.Error()
		s.emitEvent("backup:failed", eventData)
		return
	}

	s.emitEvent("backup:completed", eventData)

	// Cleanup old backups (async - sonraki backup'i bloklamamali)
	go s.backupService.CleanupOldBackups(*job, *st)
}

// GetNextRun returns the next run time for a job
func (s *Scheduler) GetNextRun(jobID int64) *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil
	}

	nextRun, err := job.NextRun()
	if err != nil {
		return nil
	}
	return &nextRun
}
