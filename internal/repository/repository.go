package repository

import (
	"context"
	"github.com/mguley/web-scraper-v1/internal/model"
)

// FilterFunc defines a function type that takes a Job and returns a boolean value.
// This function type is used for filtering job postings based on specific criteria.
type FilterFunc func(job model.Job) bool

// JobRepository is the interface for saving jobs and retrieving them.
// It defines the contract for any job repository implementation to follow, ensuring that job postings are correctly
// stored and retrieved from the database.
type JobRepository interface {
	// SaveJob stores a job posting in the repository.
	// The method takes a context and a Job object, and returns an error if the saving fails.
	//
	// Parameters:
	// - ctx context.Context: The context to use for this operation.
	// - job model.Job: The job posting to be saved in the repository.
	//
	// Returns:
	// - error: An error object if there is a failure in saving the job, otherwise nil.
	SaveJob(ctx context.Context, job model.Job) error

	// GetJobs retrieves all job postings from the repository.
	// The method takes a context and returns a slice of Job objects and an error if the retrieval fails.
	//
	// Parameters:
	//- ctx context.Context: The context to use for this operation.
	//
	// Returns:
	// - []model.Job: A slice of Job objects retrieved from the repository.
	// - error: An error object if there is a failure in retrieving the jobs, otherwise nil.
	GetJobs(ctx context.Context) ([]model.Job, error)

	// FilterBy retrieves job postings from the repository that match the given filter criteria.
	// The method takes a context and a FilterFunc, and returns a slice of Job objects and an error if the retrieval fails.
	//
	// Parameters:
	// - ctx context.Context: The context to use for this operation.
	// - filter FilterFunc: A function that defines the criteria for filtering the job postings.
	//
	// Returns:
	// - []model.Job: A slice of Job objects that match the filter criteria.
	// - error: An error object if there is a failure in filtering the jobs, otherwise nil.
	FilterBy(ctx context.Context, filter FilterFunc) ([]model.Job, error)

	// Close gracefully closes the repository connection.
	// It ensures that all resources are properly released.
	//
	// Parameters:
	// - ctx context.Context: The context to use for this operation.
	//
	// Returns:
	// - error: An error if there is a failure in closing the connection.
	Close(ctx context.Context) error
}
