package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var dockerSetupEnvironments = []string{
	"development",
	"staging",
	"production",
}

// Main function
func main() {
	setupEnvironment, clean := cliHandler()

	// Check if the setup environment is valid
	if !isInSetupList(setupEnvironment) {
		fmt.Printf("Error: The docker setup environment provided in --setup is not valid.\n")
		os.Exit(1)
	}

	fmt.Printf("Running setup environment: %s (clean: %v)\n", setupEnvironment, clean)

	// Execute the appropriate command based on the setup environment
	cmd, err := cmdHandler(setupEnvironment, clean)
	if err != nil {
		fmt.Printf("Error creating command: %v\n", err)
		os.Exit(1)
	}

	// Execute the command and handle errors
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %v\nOutput: %s\n", err, output)
		os.Exit(1)
	}

	// Print the command output
	fmt.Println(string(output))
}

func cliHandler() (string, bool) {
	// Define flags for setup and clean
	var setupEnvironment string
	var clean bool

	flag.StringVar(&setupEnvironment, "setup", "", "Name of the docker setup (e.g., development, staging, production).")
	flag.BoolVar(&clean, "clean", false, "Clean entire local development environment and start from scratch.")

	// Customize the help message
	flag.Usage = func() {
		fmt.Println("Usage: docker_setup [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  --setup [environment]     Name of the docker setup (e.g., development, staging, production).")
		fmt.Println("  --clean                   Clean entire local development environment and start from scratch.")
		fmt.Println("  --help                    Display this help message.")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println("  development:             Sets up the development environment.")
		fmt.Println("  staging:                 Sets up the staging environment.")
		fmt.Println("  production:              Sets up the production environment.")
		fmt.Println()
		os.Exit(0)
	}

	// Parse the flags
	flag.Parse()

	// Check if the help flag was passed, and call the custom usage function if necessary
	helpFlag := flag.Lookup("help")
	if helpFlag != nil && helpFlag.Value.String() == "true" {
		flag.Usage()
	}

	return setupEnvironment, clean
}

func isInSetupList(cliArgument string) bool {
	for _, env := range dockerSetupEnvironments {
		if env == cliArgument {
			return true
		}
	}
	return false
}

// Function to handle command execution based on setup environment
func cmdHandler(cliSetupArgument string, clean bool) (*exec.Cmd, error) {
	dockerComposeFilePath := "../docker-compose-" + cliSetupArgument + ".yml"

	var cmd *exec.Cmd

	// Handle clean setup
	if clean && cliSetupArgument == "development" {
		// Create the command for clean setup
		cmd = exec.Command("docker", "compose", "-f", dockerComposeFilePath, "build", "--no-cache")

		// Execute the clean setup command
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("error executing clean setup command: %v", err)
		}

		// Create the command for `docker compose up -d` if the clean setup was successful
		cmd = exec.Command("docker", "compose", "-f", dockerComposeFilePath, "up", "-d")
		return cmd, nil
	}

	// Create the command for regular setup
	cmd = exec.Command("docker", "compose", "-f", dockerComposeFilePath, "up", "-d", "--build")

	return cmd, nil
}
