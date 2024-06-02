package integration

import (
	"context"
	"flag"
	"github.com/google/uuid"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/mguley/web-scraper-v1/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"testing"
	"time"
)

const (
	timeout = time.Second * 10
)

var (
	envPath string
)

func init() {
	// Initialize flags
	flag.StringVar(&envPath, "env", "config/.env", "path to env file")
}

// setupDB initializes the database configuration and establishes a connection to MongoDB.
// It returns a MongoRepository instance or an error if the connection or configuration fails.
//
// Returns:
// - *repository.MongoRepository: A pointer to the MongoRepository instance.
// - error: An error if there is an issue with configuration or connection to MongoDB.
func setupDB() (*repository.MongoRepository, error) {
	// Initialize configuration
	if initConfigErr := config.InitConfig(envPath); initConfigErr != nil {
		log.Printf("Failed to initialize configuration: %v", initConfigErr)
		return nil, initConfigErr
	}

	appConfig, configErr := config.GetConfig()
	if configErr != nil {
		log.Printf("Failed to get configuration: %v", configErr)
		return nil, configErr
	}

	// Create a context with timeout for connecting to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Establish connection to MongoDB using the configuration
	mongoDB, createErr := repository.NewMongoJobRepository(ctx, appConfig.MongoDB)
	if createErr != nil {
		return nil, createErr
	}

	return mongoDB, nil
}

// cleanUpDB clears the MongoDB collection and closes the MongoDB connection.
// It ensures the database is cleaned up and the connection is closed after each test.
//
// Parameters:
// - mongoDB *repository.MongoRepository: The MongoRepository instance to be cleaned up.
func cleanUpDB(mongoDB *repository.MongoRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Clear the collection after each test
	_, err := mongoDB.Collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to clear the collection: %v", err)
	}

	if closeErr := mongoDB.Close(ctx); closeErr != nil {
		log.Printf("Failed to close MongoDB connection: %v", closeErr)
	}
}

// TestSaveJob tests the SaveJob function of the MongoRepository.
// It verifies that a job can be saved to the MongoDB collection without errors.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestSaveJob(t *testing.T) {
	mongoDB, err := setupDB()
	if err != nil {
		t.Errorf("Failed to setup DB: %v", err)
	}
	defer cleanUpDB(mongoDB)

	job := model.Job{
		ID:          uuid.NewString(),
		Company:     "company",
		Title:       "Software Developer",
		Description: `This is a software developer job.`,
		URL:         "https://www.example.com",
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Test the SaveJob function
	saveErr := mongoDB.SaveJob(ctx, job)
	assert.NoError(t, saveErr, "Error saving a job to the DB.")
}

// TestGetJobs tests the GetJobs function of the MongoRepository.
// It verifies that jobs can be retrieved from the MongoDB collection without errors.
// It also checks that the number of jobs retrieved matches the number of jobs saved.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestGetJobs(t *testing.T) {
	mongoDB, err := setupDB()
	if err != nil {
		t.Errorf("Failed to setup DB: %v", err)
	}
	defer cleanUpDB(mongoDB)

	jobs := []model.Job{
		{ID: uuid.NewString(), Company: "company1", Title: "Job 1", Description: "Description 1", URL: "https://www.example.com/1"},
		{ID: uuid.NewString(), Company: "company2", Title: "Job 2", Description: "Description 2", URL: "https://www.example.com/2"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, job := range jobs {
		saveErr := mongoDB.SaveJob(ctx, job)
		assert.NoError(t, saveErr, "Error saving job to the DB.")
	}

	retrievedJobs, getErr := mongoDB.GetJobs(ctx)
	assert.NoError(t, getErr, "Error retrieving jobs from the DB.")
	assert.Len(t, retrievedJobs, len(jobs), "The number of jobs retrieved does not match the number of jobs saved.")
}

// TestFilterBy tests the FilterBy function of the MongoRepository.
// It verifies that jobs can be filtered from the MongoDB collection based on specific criteria without errors.
// It also checks that the number of filtered jobs matches the expected number.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestFilterBy(t *testing.T) {
	mongoDB, err := setupDB()
	if err != nil {
		t.Errorf("Failed to setup DB: %v", err)
	}
	defer cleanUpDB(mongoDB)

	jobs := []model.Job{
		{ID: uuid.NewString(), Company: "company1", Title: "Job 1", Description: "Description 1", URL: "https://www.example.com/1"},
		{ID: uuid.NewString(), Company: "company2", Title: "Job 2", Description: "Description 2", URL: "https://www.example.com/2"},
		{ID: uuid.NewString(), Company: "company1", Title: "Job 3", Description: "Description 3", URL: "https://www.example.com/3"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, job := range jobs {
		saveErr := mongoDB.SaveJob(ctx, job)
		assert.NoError(t, saveErr, "Error saving job to the DB.")
	}

	filterFunc := func(job model.Job) bool {
		return job.Company == "company1"
	}

	filteredJobs, filterErr := mongoDB.FilterBy(ctx, filterFunc)
	assert.NoError(t, filterErr, "Error filtering by the job list.")
	assert.Len(t, filteredJobs, 2, "The number of filtered jobs does not match the expected number.")
}
