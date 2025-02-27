package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jasonlovesdoggo/caddy-defender/ranges/data"
	"github.com/jasonlovesdoggo/caddy-defender/ranges/fetchers"
	"github.com/jasonlovesdoggo/caddy-defender/ranges/fetchers/aws"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"
)

var (
	outputFormat string
	outputFile   string
)

func main() {
	// Define flags
	flag.StringVar(&outputFormat, "format", "json", "Output format: json or go")
	flag.StringVar(&outputFile, "output", "output.json", "Output file path")
	flag.Parse()

	// Create an array of all IP range fetchers
	fetchersList := []fetchers.IPRangeFetcher{
		fetchers.OpenAIFetcher{},                  // OpenAI services
		fetchers.DeepSeekFetcher{},                // DeepSeek
		fetchers.GithubCopilotFetcher{},           // GitHub Copilot
		fetchers.AzurePublicCloudFetcher{},        // Azure Public Cloud
		fetchers.GCloudFetcher{},                  // Google Cloud Platform
		aws.AWSFetcher{},                          // Global AWS IP ranges
		aws.AWSRegionFetcher{Region: "us-east-1"}, // us-east-1 region
		aws.AWSRegionFetcher{Region: "us-west-1"}, // us-west-1 region
		aws.AWSRegionFetcher{Region: "eu-west-1"}, // eu-west-1 region
		fetchers.PrivateFetcher{},
	}

	// Load the existing IP ranges from the data package
	ipRanges := data.IPRanges

	// Use a WaitGroup to wait for all fetchers to complete
	var wg sync.WaitGroup
	wg.Add(len(fetchersList))

	// Use a mutex to safely update the ipRanges map
	var mu sync.Mutex

	// Start fetching IP ranges concurrently
	for _, fetcher := range fetchersList {
		go func(f fetchers.IPRangeFetcher) {
			defer wg.Done()

			// Print the start of the fetching process
			fmt.Printf("🚀 Starting %s: %s\n", f.Name(), f.Description())

			// Fetch the IP ranges
			ranges, err := f.FetchIPRanges()
			if err != nil {
				fmt.Printf("❌ Error fetching %s: %v\n", f.Name(), err)
				return
			}

			// Update the map with the fetched ranges
			mu.Lock()
			ipRanges[strings.ToLower(f.Name())] = ranges
			mu.Unlock()

			// Print the completion of the fetching process
			fmt.Printf("✅ Completed %s: Fetched %d IP ranges\n", f.Name(), len(ranges))
		}(fetcher)
	}

	wg.Wait()

	// Handle output based on the format flag
	switch outputFormat {
	case "json":
		writeJSON(ipRanges, outputFile)
	case "go":
		outputFile = strings.Replace(outputFile, ".json", ".go", 1)
		writeGoFile(ipRanges, outputFile)
	default:
		log.Fatalf("Invalid output format: %s. Use 'json' or 'go'", outputFormat)
	}

	fmt.Printf("\n🎉 All IP ranges have been successfully written to %s\n", outputFile)
}

// writeJSON writes the IP ranges to a JSON file.
func writeJSON(ipRanges map[string][]string, outputFile string) {
	jsonData, err := json.MarshalIndent(ipRanges, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal IP ranges to JSON: %v", err)
	}

	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON to file: %v", err)
	}
}

// writeGoFile writes the IP ranges to a Go file.
func writeGoFile(ipRanges map[string][]string, outputFile string) {
	const goTemplate = `package data

// Code generated by github.com/JasonLovesDoggo/caddy-defender/blob/main/ranges/main.go; DO NOT EDIT.

var IPRanges = map[string][]string{
	{{- range $key, $values := . }}
	"{{ $key }}": { {{- range $index, $value := $values }}
		"{{ $value }}",{{- end }}
	},{{- end }}
}
`

	// Write the generated Go file
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error closing file: %v", err)
		}
	}(file)

	t := template.Must(template.New("code").Parse(goTemplate))
	err = t.Execute(file, ipRanges)
	if err != nil {
		log.Panicf("Failed to execute template: %v", err)
	}
}
