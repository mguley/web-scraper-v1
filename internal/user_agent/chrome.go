package user_agent

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ChromeUserAgentGenerator implements the UserAgentGenerator interface specifically for Google Chrome browsers.
// It generates User-Agent strings that simulate various versions of Chrome on different operating systems.
//
// Fields:
// - versions: A slice of strings representing different versions of the Chrome browser.
// - operatingSystems: A slice of strings representing different operating systems.
// - randSource: A pointer to a rand.Rand instance used to generate random numbers for selecting versions and operating systems.
//
// Methods:
// - NewChromeUserAgentGenerator: Initializes a new instance of ChromeUserAgentGenerator with up-to-date data.
// - Generate: Constructs a User-Agent string simulating a Chrome browser on various operating systems.
type ChromeUserAgentGenerator struct {
	versions         []string
	operatingSystems []string
	randSource       *rand.Rand
}

var (
	randInstance *rand.Rand
	once         sync.Once
)

// initRand initializes the randInstance singleton.
// It ensures that the random source is seeded only once to maintain thread safety and avoid repeated initialization.
func initRand() {
	once.Do(func() {
		randInstance = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	})
}

// NewChromeUserAgentGenerator initializes a new instance of ChromeUserAgentGenerator with up-to-date data.
//
// Returns:
// - *ChromeUserAgentGenerator: A pointer to a ChromeUserAgentGenerator instance.
func NewChromeUserAgentGenerator() *ChromeUserAgentGenerator {
	initRand()
	return &ChromeUserAgentGenerator{
		versions: []string{
			"126.0.6478.114", "126.0.6478.62", "126.0.6478.61",
			"126.0.6478.56", "124.0.6367.243", "124.0.6367.233",
			"124.0.6367.230", "124.0.6367.221", "124.0.6367.208",
			"124.0.6367.201", "124.0.6367.118", "123.0.6358.132",
			"123.0.6358.121", "122.0.6345.98", "122.0.6345.67",
		},
		operatingSystems: []string{
			"Windows NT 10.0; Win64; x64",
			"Macintosh; Intel Mac OS X 10_15_7",
			"X11; Linux x86_64", "Windows NT 6.1; Win64; x64",
			"Macintosh; Intel Mac OS X 10_14_6",
		},
		randSource: randInstance,
	}
}

// Generate constructs a User-Agent string simulating a Chrome browser on various operating systems.
// It randomly selects a version and an operating system from the pre-defined lists and formats them into a User-Agent string.
//
// Returns:
// - string: A User-Agent string representing a specific version of the Chrome browser on a specific operating system.
func (chrome *ChromeUserAgentGenerator) Generate() string {
	version := chrome.versions[chrome.randSource.Intn(len(chrome.versions))]
	os := chrome.operatingSystems[chrome.randSource.Intn(len(chrome.operatingSystems))]

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", os, version)
}
