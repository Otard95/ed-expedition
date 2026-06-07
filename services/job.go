package services

import (
	"context"
	"ed-expedition/lib/channels"
	"ed-expedition/lib/job"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type jobEntry interface {
	Status() job.JobStatus
	Start(ctx context.Context)
	StatusChange() *channels.FanoutChannel[job.JobStatus]
}

type JobService struct {
	jobs   map[string]jobEntry
	logger wailsLogger.Logger

	JobStatus *channels.FanoutChannel[*job.JobStatus]
}

func NewJobService(logger wailsLogger.Logger) *JobService {
	return &JobService{
		jobs:      make(map[string]jobEntry, 8),
		JobStatus: channels.NewFanoutChannel[*job.JobStatus]("JobStatus", 0, time.Millisecond, logger),
		logger:    logger,
	}
}

func (j *JobService) Start() {}

func (j *JobService) Stop() {
	j.JobStatus.Close()
}

func (j *JobService) RegisterJob(id string, job jobEntry) {
	j.jobs[id] = job

	go func() {
		for status := range job.StatusChange().Subscribe() {
			j.JobStatus.Publish(&status)
		}
	}()
}

func (j *JobService) RegisterAndRun(entry jobEntry, ctx context.Context) {
	id := entry.Status().ID
	j.RegisterJob(id, entry)
	go entry.Start(ctx)
}
