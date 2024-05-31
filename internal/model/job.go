package model

import (
	"fmt"
)

// Job represents a job posting with various details such as title, company, description, and URL.
type Job struct {
	ID          string `bson:"_id,omitempty" json:"id,omitempty"` // ID: A unique identifier for the job posting.
	Title       string `bson:"title" json:"title"`                // Title: The title of the job position.
	Company     string `bson:"company" json:"company"`            // Company: The name of the company offering the job.
	Description string `bson:"description" json:"description"`    // Description: A detailed description of the job responsibilities and requirements.
	URL         string `bson:"url" json:"url"`                    // URL: The URL to the original job posting.
}

// String returns a string representation of the Job struct.
//
// Returns:
// - A string that includes the job's ID, title, company, description, and URL.
func (job *Job) String() string {
	return fmt.Sprintf(
		"Job{ID: %s, Title: %s, Company: %s, Description: %s, URL: %s}",
		job.ID,
		job.Title,
		job.Company,
		job.Description,
		job.URL,
	)
}
