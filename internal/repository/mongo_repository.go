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
	"time"
)

const (
	Timeout = 10 * time.Second
)

// MongoRepository implements the JobRepository interface.
// It provides functionality to interact with MongoDB for job data.
type MongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoJobRepository creates a new instance of MongoRepository.
// It connects to MongoDB using the provided configuration settings.
//
// Parameters:
// - db: The MongoDB configuration settings.
//
// Returns:
// - *MongoRepository: A pointer to an instance of MongoRepository.
// - error: An error if there is an issue establishing the connection or creating the collection.
func NewMongoJobRepository(db config.MongoDB) (*MongoRepository, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", db.User, db.Pass, db.Host, db.Port)
	clientOptions := options.Client().ApplyURI(uri)

	// Create a new context for the connection
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		utils.GetLogger().LogError(err, true)
		return nil, err
	}

	collection := client.Database(db.Database).Collection(db.Collection)
	return &MongoRepository{client: client, collection: collection}, nil
}

// SaveJob saves a job to the MongoDB collection.
//
// Parameters:
// - job model.Job: The job to be saved.
//
// Returns:
// - error: An error if there is an issue saving the job.
func (mongoRepository *MongoRepository) SaveJob(job model.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	_, err := mongoRepository.collection.InsertOne(ctx, job)
	if err != nil {
		utils.GetLogger().LogError(err, true)
		return err
	}

	utils.GetLogger().LogInfo("Job Inserted: " + job.Title)
	return nil
}

// GetJobs retrieves all jobs from the MongoDB collection.
//
// Returns:
// - []model.Job: A slice of jobs.
// - error: An error if there is an issue retrieving the jobs.
func (mongoRepository *MongoRepository) GetJobs() ([]model.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	cursor, findErr := mongoRepository.collection.Find(ctx, bson.M{})
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
	for cursor.Next(ctx) {
		var item model.Job
		if decodeErr := cursor.Decode(&item); decodeErr != nil {
			utils.GetLogger().LogError(decodeErr, true)
			continue
		}
		items = append(items, item)
	}

	if cursorErr := cursor.Err(); cursorErr != nil {
		utils.GetLogger().LogError(cursorErr, true)
		return nil, cursorErr
	}

	utils.GetLogger().LogInfo("Retrieved jobs from the database.")
	return items, nil
}

// FilterBy retrieves job postings from the MongoDB that match the given filter criteria.
//
// Parameters:
// - filter FilterFunc: A function that defines the criteria for filtering the job postings.
//
// Returns:
// - []model.Job: A slice of Job objects that match the filter criteria.
// - error: An error object if there is a failure in filtering the jobs, otherwise nil.
func (mongoRepository *MongoRepository) FilterBy(filter FilterFunc) ([]model.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	cursor, findErr := mongoRepository.collection.Find(ctx, bson.M{})
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
	for cursor.Next(ctx) {
		var item model.Job
		if decodeErr := cursor.Decode(&item); decodeErr != nil {
			utils.GetLogger().LogError(decodeErr, true)
			continue
		}
		if filter(item) {
			items = append(items, item)
		}
	}

	if cursorErr := cursor.Err(); cursorErr != nil {
		utils.GetLogger().LogError(cursorErr, true)
		return nil, cursorErr
	}

	utils.GetLogger().LogInfo("Filtered jobs from the database.")
	return items, nil
}

// Close gracefully closes the MongoDB client connection.
// It ensures that all resources are properly released.
//
// Returns:
// - error: An error if there is an issue closing the connection.
func (mongoRepository *MongoRepository) Close() error {
	disconnectErr := mongoRepository.client.Disconnect(context.TODO())
	if disconnectErr != nil {
		utils.GetLogger().LogError(disconnectErr, true)
		return disconnectErr
	}
	return nil
}
