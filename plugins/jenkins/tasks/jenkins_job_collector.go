package tasks

import (
	"context"
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/merico-dev/lake/logger"
	lakeModels "github.com/merico-dev/lake/models"
	"github.com/merico-dev/lake/plugins/jenkins/models"
	"github.com/merico-dev/lake/utils"
	"gorm.io/gorm/clause"
)

func CollectJobs(apiClient *JenkinsApiClient, scheduler *utils.WorkerScheduler, ctx context.Context) error {
	// get all jobs
	var jobs, err = apiClient.jenkins.GetAllJobs(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get jobs from jenkins: %v", err)
	}

	for _, v := range jobs {
		logger.Debug("(collect job) Submit", v)
		workerErr := syncJob(v)
		if workerErr != nil {
			return workerErr
		}
	}
	scheduler.WaitUntilFinish()
	return nil
}

func syncJob(job *gojenkins.Job) error {
	logger.Info("syncJob", job.Raw.Name)
	var jenkinsJob = models.JenkinsJob{
		JenkinsJobProps: models.JenkinsJobProps{
			Name:  job.Raw.Name,
			Class: job.Raw.Class,
			Color: job.Raw.Color,
		},
	}
	err := lakeModels.Db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&jenkinsJob).Error

	if err != nil {
		return fmt.Errorf("failed to save job: %v", err)
	}
	return nil
}