package main

import (
	"crownet/cli"
	"crownet/config"
	"fmt"
	"log"
	"os"
)

func main() {
	// Load application configuration. This includes parsing CLI flags
	// and validating the overall configuration.
	appCfg, err := config.NewAppConfig()
	if err != nil {
		// log.Fatalf will print to stderr and exit with status 1.
		// This is suitable for fatal configuration errors.
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize the CLI orchestrator with the application configuration.
	orchestrator := cli.NewOrchestrator(appCfg)

	// Run the main application logic via the orchestrator.
	if err := orchestrator.Run(); err != nil {
		// For errors during the orchestrator's run, printing to Stderr
		// without the log package's timestamp/prefix can be cleaner for user output.
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		os.Exit(1) // Indicate an error occurred.
	}

	// If orchestrator.Run() completes without error, exit successfully.
	// fmt.Println("CrowNet finished successfully.") // Optional success message
}
