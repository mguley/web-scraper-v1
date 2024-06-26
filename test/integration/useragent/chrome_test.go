package useragent

import (
	"fmt"
	"github.com/mguley/web-scraper-v1/internal/useragent"
	"github.com/stretchr/testify/require"
	"regexp"
	"sync"
	"testing"
)

// TestChromeUserAgentGeneratorBasicFunctionality verifies that the User-Agent generator can produce a valid,
// non-empty User-Agent string.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChromeUserAgentGeneratorBasicFunctionality(t *testing.T) {
	generator := useragent.NewChromeUserAgentGenerator()
	require.NotNil(t, generator)

	ua := generator.Generate()
	require.NotEmpty(t, ua, "Generated User-Agent string should not be empty")
}

// TestChromeUserAgentGeneratorDiversity ensures that the User-Agent generator produces a diverse set of User-Agent
// strings, testing both the diversity in versions and operating systems.
// It checks for a reasonable spread of different Chrome versions and operating systems to simulate varying user environments.
//
// Parameters:
// - t *testing.T: The testing framework's instance.
func TestChromeUserAgentGeneratorDiversity(t *testing.T) {
	generator := useragent.NewChromeUserAgentGenerator()
	require.NotNil(t, generator, "Generator should be properly initialized")

	versionsSeen := make(map[string]bool)
	osSeen := make(map[string]bool)
	numSamples := 20 // Number of samples to generate for diversity test

	for i := 0; i < numSamples; i++ {
		ua := generator.Generate()
		version, os, err := parseUserAgent(ua)
		require.NoError(t, err, "User-Agent string should follow the expected format")
		versionsSeen[version] = true
		osSeen[os] = true
	}

	require.GreaterOrEqual(t, len(versionsSeen), 5, "Expected at least 5 different versions")
	require.GreaterOrEqual(t, len(osSeen), 3, "Expected at least 3 different operating systems")
}

// TestChromeUserAgentGeneratorConcurrency verifies that the User-Agent generator maintains correct functionality
// and data integrity under high concurrency. The test simulates a scenario where multiple goroutines concurrently
// generate User-Agent strings to ensure that the generator can handle parallel operations without producing any errors.
//
// The test divides the task among several workers (goroutines), each responsible for generating a segment of the total
// number of User-Agent strings. It confirms that the sum of the generated strings matches the expected total and that
// all strings are valid and non-empty, assessing the generator's readiness for concurrent use in production environments.
//
// Parameters:
// - t *testing.T: The testing framework's instance,
func TestChromeUserAgentGeneratorConcurrency(t *testing.T) {
	generator := useragent.NewChromeUserAgentGenerator()
	require.NotNil(t, generator)

	numSamples := 100_000
	numWorkers := 10
	samplesPerWorker := numSamples / numWorkers
	lock := sync.Mutex{}
	generatedUserAgents := 0

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < samplesPerWorker; i++ {
				ua := generator.Generate()
				require.NotEmpty(t, ua, "Generated User-Agent string should not be empty")
				lock.Lock()
				generatedUserAgents++
				lock.Unlock()
			}
		}()
	}

	wg.Wait()

	require.Equal(t, numSamples, generatedUserAgents)
}

// parseUserAgent extracts the version and operating system from a given User-Agent string.
// The function uses a regular expression to specifically parse strings formatted to match the conventional structure
// of a Chrome browser User-Agent. It isolates and returns the operating system and Chrome version as separate strings,
// enabling further analysis or validation in other parts of the test suite.
//
// Parameters:
// - ua string: The User-Agent string to be parsed.
//
// Returns:
// - version string: The version of the Chrome browser extracted from the User-Agent string.
// - os string: The operating system information extracted from the User-Agent string.
// - err error: An error object that is non-nil if the parsing fails to match the expected format.
func parseUserAgent(ua string) (version string, os string, err error) {
	re := regexp.MustCompile(`Mоzillа/5.0 \(([^)]+)\) AppleWebKit/537.36 \(KHTML, like Gecko\) Chrome/([\d.]+) Safari/537.36`)
	matches := re.FindStringSubmatch(ua)

	if len(matches) < 3 {
		return "", "", fmt.Errorf("invalid User-Agent string format")
	}

	os = matches[1]
	version = matches[2]

	return version, os, nil
}
