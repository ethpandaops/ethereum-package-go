package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethpandaops/ethereum-package-go"
)

func main() {
	ctx := context.Background()

	// Example 1: Using default pinned version
	fmt.Println("=== Using Default Pinned Version ===")
	fmt.Printf("Default repository: %s\n", ethereum.DefaultPackageRepository)
	fmt.Printf("Default version: %s\n", ethereum.DefaultPackageVersion)

	// Example 2: Override version while keeping default repository
	fmt.Println("\n=== Override Version Only ===")
	network1, err := ethereum.Run(ctx,
		ethereum.Minimal(),
		ethereum.WithPackageVersion("2.8.0"), // Use a different version
		ethereum.WithDryRun(true),             // Just validate, don't run
	)
	if err != nil {
		log.Printf("Error with version override: %v", err)
	} else {
		fmt.Printf("Created network with custom version: %v\n", network1 != nil)
	}

	// Example 3: Override both repository and version
	fmt.Println("\n=== Override Repository and Version ===")
	network2, err := ethereum.Run(ctx,
		ethereum.Minimal(),
		ethereum.WithPackageRepo("github.com/my-org/custom-ethereum-package", "1.5.0"),
		ethereum.WithDryRun(true), // Just validate, don't run
	)
	if err != nil {
		log.Printf("Error with repo override: %v", err)
	} else {
		fmt.Printf("Created network with custom repo and version: %v\n", network2 != nil)
	}

	// Example 4: Override repository only (no version pinning)
	fmt.Println("\n=== Override Repository Only ===")
	network3, err := ethereum.Run(ctx,
		ethereum.Minimal(),
		ethereum.WithPackageID("github.com/my-org/ethereum-package-fork"),
		ethereum.WithPackageVersion(""), // Clear the version to use latest
		ethereum.WithDryRun(true),       // Just validate, don't run
	)
	if err != nil {
		log.Printf("Error with repo only override: %v", err)
	} else {
		fmt.Printf("Created network with custom repo (latest): %v\n", network3 != nil)
	}

	fmt.Println("\n=== Version Configuration Summary ===")
	fmt.Printf("✓ Default: Uses %s@%s\n", ethereum.DefaultPackageRepository, ethereum.DefaultPackageVersion)
	fmt.Println("✓ WithPackageVersion(): Override version only")
	fmt.Println("✓ WithPackageRepo(): Override both repository and version")
	fmt.Println("✓ WithPackageID(): Override repository only")
	fmt.Println("✓ Combine WithPackageID() + WithPackageVersion() for full control")
}