package repository

import "github.com/mguley/web-scraper-v1/internal/model"

// FilterFunc defines a function type that takes a Job and returns a boolean value.
// This function type is used for filtering job postings based on specific criteria.
type FilterFunc func(job model.Job) bool

// JobRepository is the interface for saving jobs and retrieving them.
// It defines the contract for any job repository implementation to follow, ensuring that job postings are correctly
// stored and retrieved from the database.
type JobRepository interface {
	// SaveJob stores a job posting in MongoDB.
	// The method takes a Job object and returns an error if the saving fails.
	//
	// Parameters:
	// - job: model.Job: The job posting to be saved in the repository.
	//
	// Returns:
	// - error: An error object if there is a failure in saving the job, otherwise nil.
	SaveJob(job model.Job) error

	// GetJobs retrieves all job postings from MongoDB.
	// The method returns a slice of Job objects and an error if the retrieval fails.
	//
	// Returns:
	// - ([]model.Job): A slice of Job objects retrieved from MongoDB.
	// - error: An error object if there is a failure in retrieving the jobs, otherwise nil.
	GetJobs() ([]model.Job, error)

	// FilterBy retrieves job postings from the MongoDB that match the given filter criteria.
	// The method takes a FilterFunc and returns a slice of Job objects and an error if the retrieval fails.
	//
	// Parameters:
	// - filter: FilterFunc: A function that defines the criteria for filtering the job postings.
	//
	// Returns:
	// - ([]model.Job): A slice of Job objects that match the filter criteria.
	// - error: An error object if there is a failure in filtering the jobs, otherwise nil.
	FilterBy(filter FilterFunc) ([]model.Job, error)
}
