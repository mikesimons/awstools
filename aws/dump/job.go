package main

import (
	"errors"
	"fmt"

	"github.com/fatih/structs"
	"github.com/hamstah/awstools/common"
)

type Resource struct {
	ID        string                 `json:"id"`
	ARN       string                 `json:"arn"`
	Service   string                 `json:"service"`
	Type      string                 `json:"type"`
	AccountID string                 `json:"account_id"`
	Region    string                 `json:"region"`
	Metadata  map[string]interface{} `json:"metadata"`
	ManagedBy map[string]string      `json:"managed_by"`
}

func (r *Resource) UniqueID() string {
	if r.ARN == "" {
		return r.ID
	}

	return r.ARN
}

func NewResource(arn string, metadata interface{}) (*Resource, error) {
	parsed, err := common.ParseARN(arn)
	if err != nil {
		return nil, err
	}

	return &Resource{
		ID:        parsed.Resource,
		ARN:       arn,
		Service:   parsed.Service,
		Type:      parsed.ResourceType,
		AccountID: parsed.AccountID,
		Region:    parsed.Region,
		Metadata:  structs.Map(metadata),
	}, nil
}

type Service struct {
	Name     string
	IsGlobal bool
	Reports  map[string]Report
}

func (s *Service) GenerateAllJobs(account *Account) ([]Job, error) {
	jobs := []Job{}
	for resource, _ := range s.Reports {
		newJobs, err := s.GenerateJobs(account, resource)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, newJobs...)
	}
	return jobs, nil
}

func (s *Service) GenerateJobs(account *Account, resource string) ([]Job, error) {
	Report, ok := s.Reports[resource]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unknown resource %s for service %s", resource, s.Name))
	}
	jobs := []Job{}
	if s.IsGlobal {
		jobs = append(jobs, Job{
			Report:  Report,
			Session: account.Sessions[0],
		})
	} else {
		for _, session := range account.Sessions {
			jobs = append(jobs, Job{
				Report:  Report,
				Session: session,
			})
		}
	}
	return jobs, nil
}

type ReportResult struct {
	Resources []Resource
	Error     error
}

type Report func(*Session) *ReportResult

type Job struct {
	Report  Report
	Session *Session
}

func worker(id int, jobs <-chan Job, results chan<- *ReportResult) {
	for job := range jobs {
		results <- job.Report(job.Session)
	}
}

func Run(jobs []Job) []Resource {
	resources := []Resource{}

	jobsChan := make(chan Job, len(jobs))
	results := make(chan *ReportResult, len(jobs))

	for w := 0; w < 10; w++ {
		go worker(w, jobsChan, results)
	}

	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)

	for i := 0; i < len(jobs); i++ {
		result := <-results
		if result.Error == nil {
			resources = append(resources, result.Resources...)
		}
	}
	return resources
}
