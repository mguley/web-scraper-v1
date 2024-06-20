package user_agent

// UserAgentGenerator defines the interface for generating User-Agent strings.
// It abstracts the process of creating user-agent strings for different browsers and devices,
// ensuring the flexibility and variability needed for effective web scraping.
//
// Methods:
// - Generate string: Generates a User-Agent string based on the implemented criteria.
type UserAgentGenerator interface {
	// Generate creates a User-Agent string.
	// It returns a user-agent string that can be used in HTTP headers to mimic requests from various browsers and devices.
	//
	// Returns:
	// - string: A User-Agent string representing a specific browser and device configuration.
	Generate() string
}
