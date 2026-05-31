package main

import "gopher/internal"

import "fmt"

func main() {
	cfg := internal.ParseClI()
	fmt.Printf("Introspecting %s with %d workers. Verbose: %t\n", cfg.Url, cfg.Workers, cfg.Verbose)

	err := internal.ValidateCLI(cfg)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	// Build the URL map for the given URL.
	urlMap := internal.BuildUrlMap(cfg.Url, nil)

	// Print the URL map in a readable format.
	internal.PrintURLMap(urlMap, 0)
}
