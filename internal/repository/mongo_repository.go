package repository

import (
	"context"
	"fmt"
	"github.com/mguley/web-scraper-v1/config"
	"github.com/mguley/web-scraper-v1/internal/model"
	"github.com/mguley/web-scraper-v1/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository implements the JobRepository interface.
// It provides functionality to interact with MongoDB for job data.
type MongoRepository struct {
	client     *mongo.Client
	Collection *mongo.Collection
}

// NewMongoJobRepository creates a new instance of MongoRepository.
// It connects to MongoDB using the provided configuration settings.
//
// Parameters:
// - ctx context.Context: The context to use for this operation.
// - db config.MongoDB: The MongoDB configuration settings.
//
// Returns:
// - *MongoRepository: A pointer to an instance of MongoRepository.
// - error: An error if there is an issue establishing the connection or creating the collection.
func NewMongoJobRepository(ctx context.Context, db config.MongoDB) (*MongoRepository, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", db.User, db.Pass, db.Host, db.Port)
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		utils.GetLogger().LogError(err, true)
		return nil, err
	}

	collection := client.Database(db.Database).Collection(db.Collection)
	return &MongoRepository{client: client, Collection: collection}, nil
}

// SaveJob saves a job to the MongoDB collection.
//
// Parameters:
// - ctx context.Context: The context to use for this operation.
// - job model.Job: The job to be saved.
//
// Returns:
// - error: An error if there is an issue saving the job.
func (mongoRepository *MongoRepository) SaveJob(ctx context.Context, job model.Job) error {
	if _, err := mongoRepository.Collection.InsertOne(ctx, job); err != nil {
		utils.GetLogger().LogError(err, true)
		return err
	}

	utils.GetLogger().LogInfo("Job Saved: " + job.String())
	return nil
}

// GetJobs retrieves all jobs from the MongoDB collection.
//
// Parameters:
// - ctx context.Context: The context to use for this operation.
//
// Returns:
// - []model.Job: A slice of jobs.
// - error: An error if there is an issue retrieving the jobs.
func (mongoRepository *MongoRepository) GetJobs(ctx context.Context) ([]model.Job, error) {
	cursor, findErr := mongoRepository.Collection.Find(ctx, bson.M{})
	if findErr != nil {
		utils.GetLogger().LogError(findErr, true)
		return nil, findErr
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			utils.GetLogger().LogError(err, true)
			return
		}
	}()

	var items []model.Job
	if getAllErr := cursor.All(ctx, &items); getAllErr != nil {
		utils.GetLogger().LogError(getAllErr, true)
		return nil, getAllErr
	}

	message := fmt.Sprintf("Retrieved %d jobs", len(items))
	utils.GetLogger().LogInfo(message)
	return items, nil
}

// FilterBy retrieves job postings from the MongoDB that match the given filter criteria.
//
// Parameters:
// - ctx context.Context: The context to use for this operation.
// - filter FilterFunc: A function that defines the criteria for filtering the job postings.
//
// Returns:
// - []model.Job: A slice of Job objects that match the filter criteria.
// - error: An error object if there is a failure in filtering the jobs, otherwise nil.
func (mongoRepository *MongoRepository) FilterBy(ctx context.Context, filter FilterFunc) ([]model.Job, error) {
	items, getItemsErr := mongoRepository.GetJobs(ctx)
	if getItemsErr != nil {
		return nil, getItemsErr
	}

	var filteredJobs []model.Job
	for _, job := range items {
		if filter(job) {
			filteredJobs = append(filteredJobs, job)
		}
	}

	message := fmt.Sprintf("Filtered %d jobs", len(filteredJobs))
	utils.GetLogger().LogInfo(message)
	return filteredJobs, nil
}

// Close gracefully closes the MongoDB client connection.
// It ensures that all resources are properly released.
//
// Parameters:
// - ctx context.Context: The context to use for this operation.
//
// Returns:
// - error: An error if there is an issue closing the connection.
func (mongoRepository *MongoRepository) Close(ctx context.Context) error {
	if disconnectErr := mongoRepository.client.Disconnect(ctx); disconnectErr != nil {
		utils.GetLogger().LogError(disconnectErr, true)
		return disconnectErr
	}
	return nil
}
