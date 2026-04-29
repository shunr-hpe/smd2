package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	v1 "github.com/OpenCHAMI/inventory-service/apis/inventory-service.openchami.org/v1"
	"github.com/OpenCHAMI/inventory-service/cmd/plugins"
	"github.com/OpenCHAMI/inventory-service/internal/storage"
)

func newImportCommand() *cobra.Command {
	var (
		input        string
		mode         string
		dryRun       bool
		skipExisting bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import resources from files",
		Long: `Import resources from JSON or YAML files into storage.

This is useful for:
  - Restoring from backups
  - Migrating data between environments
  - Bulk loading resource definitions
  - Testing with known resource state

Import modes:
  - upsert: Create new resources or update existing (default)
  - replace: Delete all resources first, then import
  - skip: Skip resources that already exist

Examples:
  # Import from backup directory
  inventory_service import --input ./backup

  # Dry run to preview changes
  inventory_service import --input ./backup --dry-run

  # Replace all resources
  inventory_service import --input ./backup --mode replace
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(cmd.Context(), input, mode, dryRun, skipExisting)
		},
	}

	cmd.Flags().StringVar(&input, "input", "./backup", "Input directory containing resource files")
	cmd.Flags().StringVar(&mode, "mode", "upsert", "Import mode: upsert, replace, skip")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().BoolVar(&skipExisting, "skip-existing", false, "Skip resources that already exist (same as --mode skip)")

	return cmd
}

func runImport(ctx context.Context, input, mode string, dryRun, skipExisting bool) error {
	fmt.Printf("🚀 Importing resources...\n")
	fmt.Printf("   Input: %s\n", input)
	fmt.Printf("   Mode: %s\n", mode)
	if dryRun {
		fmt.Printf("   ⚠️  DRY RUN - No changes will be applied\n")
	}

	// Validate mode
	if skipExisting {
		mode = "skip"
	}
	if mode != "upsert" && mode != "replace" && mode != "skip" {
		return fmt.Errorf("unsupported mode: %s (use 'upsert', 'replace', or 'skip')", mode)
	}

	// Check input directory exists
	if _, err := os.Stat(input); err != nil {
		return fmt.Errorf("input directory does not exist: %w", err)
	}

	// Handle replace mode - delete all resources first
	if mode == "replace" && !dryRun {
		fmt.Printf("⚠️  Replace mode - deleting existing resources...\n")
		if err := deleteAllResources(ctx); err != nil {
			return fmt.Errorf("failed to delete existing resources: %w", err)
		}
	}

	// Walk input directory and import files
	totalImported := 0
	totalSkipped := 0
	var importErr error

	err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process JSON and YAML files
		ext := filepath.Ext(path)
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return nil
		}

		imported, skipped, err := importFile(ctx, path, mode, dryRun)
		if err != nil {
			fmt.Printf("  ✗ %s: %v\n", filepath.Base(path), err)
			importErr = err
			return nil // Continue with other files
		}
		totalImported += imported
		totalSkipped += skipped
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk import directory: %w", err)
	}

	if dryRun {
		fmt.Printf("✅ Dry run complete. Would import %d resources (%d skipped).\n", totalImported, totalSkipped)
	} else {
		fmt.Printf("✅ Import complete. Imported %d resources (%d skipped).\n", totalImported, totalSkipped)
	}

	return importErr
}

func importFile(ctx context.Context, path string, mode string, dryRun bool) (imported, skipped int, err error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine format
	ext := filepath.Ext(path)

	// Try to unmarshal into generic resource first to determine kind
	var genericResource struct {
		APIVersion string `json:"apiVersion" yaml:"apiVersion"`
		Kind       string `json:"kind" yaml:"kind"`
	}

	if ext == ".json" {
		if err := json.Unmarshal(data, &genericResource); err != nil {
			return 0, 0, fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &genericResource); err != nil {
			return 0, 0, fmt.Errorf("failed to parse YAML: %w", err)
		}
	}

	// Import based on kind
	switch genericResource.Kind {
	case "Component":
		var res *v1.Component
		if ext == ".json" {
			if err := json.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal Component: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal Component: %w", err)
			}
		}

		// Check if resource exists
		existing, err := storage.GetComponentByUID(ctx, res.Metadata.UID)
		if err == nil && existing != nil {
			// Resource exists
			if mode == "skip" {
				fmt.Printf("  ⊘ %s (exists)\n", filepath.Base(path))
				return 0, 1, nil
			}
			fmt.Printf("  ⟳ %s (updating)\n", filepath.Base(path))
		} else {
			fmt.Printf("  ✓ %s (creating)\n", filepath.Base(path))
		}

		if !dryRun {
			if err := plugins.Store.SaveComponent(ctx, res); err != nil {
				return 0, 0, fmt.Errorf("failed to save Component: %w", err)
			}
		}
		return 1, 0, nil
	case "ComponentEndpoint":
		var res *v1.ComponentEndpoint
		if ext == ".json" {
			if err := json.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal ComponentEndpoint: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal ComponentEndpoint: %w", err)
			}
		}

		// Check if resource exists
		existing, err := storage.GetComponentEndpointByUID(ctx, res.Metadata.UID)
		if err == nil && existing != nil {
			// Resource exists
			if mode == "skip" {
				fmt.Printf("  ⊘ %s (exists)\n", filepath.Base(path))
				return 0, 1, nil
			}
			fmt.Printf("  ⟳ %s (updating)\n", filepath.Base(path))
		} else {
			fmt.Printf("  ✓ %s (creating)\n", filepath.Base(path))
		}

		if !dryRun {
			if err := plugins.Store.SaveComponentEndpoint(ctx, res); err != nil {
				return 0, 0, fmt.Errorf("failed to save ComponentEndpoint: %w", err)
			}
		}
		return 1, 0, nil
	case "EthernetInterface":
		var res *v1.EthernetInterface
		if ext == ".json" {
			if err := json.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal EthernetInterface: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal EthernetInterface: %w", err)
			}
		}

		// Check if resource exists
		existing, err := storage.GetEthernetInterfaceByUID(ctx, res.Metadata.UID)
		if err == nil && existing != nil {
			// Resource exists
			if mode == "skip" {
				fmt.Printf("  ⊘ %s (exists)\n", filepath.Base(path))
				return 0, 1, nil
			}
			fmt.Printf("  ⟳ %s (updating)\n", filepath.Base(path))
		} else {
			fmt.Printf("  ✓ %s (creating)\n", filepath.Base(path))
		}

		if !dryRun {
			if err := plugins.Store.SaveEthernetInterface(ctx, res); err != nil {
				return 0, 0, fmt.Errorf("failed to save EthernetInterface: %w", err)
			}
		}
		return 1, 0, nil
	case "Group":
		var res *v1.Group
		if ext == ".json" {
			if err := json.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal Group: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal Group: %w", err)
			}
		}

		// Check if resource exists
		existing, err := storage.GetGroupByUID(ctx, res.Metadata.UID)
		if err == nil && existing != nil {
			// Resource exists
			if mode == "skip" {
				fmt.Printf("  ⊘ %s (exists)\n", filepath.Base(path))
				return 0, 1, nil
			}
			fmt.Printf("  ⟳ %s (updating)\n", filepath.Base(path))
		} else {
			fmt.Printf("  ✓ %s (creating)\n", filepath.Base(path))
		}

		if !dryRun {
			if err := plugins.Store.SaveGroup(ctx, res); err != nil {
				return 0, 0, fmt.Errorf("failed to save Group: %w", err)
			}
		}
		return 1, 0, nil
	case "Hardware":
		var res *v1.Hardware
		if ext == ".json" {
			if err := json.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal Hardware: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal Hardware: %w", err)
			}
		}

		// Check if resource exists
		existing, err := storage.GetHardwareByUID(ctx, res.Metadata.UID)
		if err == nil && existing != nil {
			// Resource exists
			if mode == "skip" {
				fmt.Printf("  ⊘ %s (exists)\n", filepath.Base(path))
				return 0, 1, nil
			}
			fmt.Printf("  ⟳ %s (updating)\n", filepath.Base(path))
		} else {
			fmt.Printf("  ✓ %s (creating)\n", filepath.Base(path))
		}

		if !dryRun {
			if err := storage.SaveHardware(ctx, res); err != nil {
				return 0, 0, fmt.Errorf("failed to save Hardware: %w", err)
			}
		}
		return 1, 0, nil
	case "RedfishEndpoint":
		var res *v1.RedfishEndpoint
		if ext == ".json" {
			if err := json.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal RedfishEndpoint: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal RedfishEndpoint: %w", err)
			}
		}

		// Check if resource exists
		existing, err := storage.GetRedfishEndpointByUID(ctx, res.Metadata.UID)
		if err == nil && existing != nil {
			// Resource exists
			if mode == "skip" {
				fmt.Printf("  ⊘ %s (exists)\n", filepath.Base(path))
				return 0, 1, nil
			}
			fmt.Printf("  ⟳ %s (updating)\n", filepath.Base(path))
		} else {
			fmt.Printf("  ✓ %s (creating)\n", filepath.Base(path))
		}

		if !dryRun {
			if err := plugins.Store.SaveRedfishEndpoint(ctx, res); err != nil {
				return 0, 0, fmt.Errorf("failed to save RedfishEndpoint: %w", err)
			}
		}
		return 1, 0, nil
	case "ServiceEndpoint":
		var res *v1.ServiceEndpoint
		if ext == ".json" {
			if err := json.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal ServiceEndpoint: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &res); err != nil {
				return 0, 0, fmt.Errorf("failed to unmarshal ServiceEndpoint: %w", err)
			}
		}

		// Check if resource exists
		existing, err := storage.GetServiceEndpointByUID(ctx, res.Metadata.UID)
		if err == nil && existing != nil {
			// Resource exists
			if mode == "skip" {
				fmt.Printf("  ⊘ %s (exists)\n", filepath.Base(path))
				return 0, 1, nil
			}
			fmt.Printf("  ⟳ %s (updating)\n", filepath.Base(path))
		} else {
			fmt.Printf("  ✓ %s (creating)\n", filepath.Base(path))
		}

		if !dryRun {
			if err := plugins.Store.SaveServiceEndpoint(ctx, res); err != nil {
				return 0, 0, fmt.Errorf("failed to save ServiceEndpoint: %w", err)
			}
		}
		return 1, 0, nil
	default:
		return 0, 0, fmt.Errorf("unknown resource kind: %s", genericResource.Kind)
	}
}

func deleteAllResources(ctx context.Context) error {
	// Delete all components
	componentItems, err := storage.Querycomponents(ctx).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query components: %w", err)
	}
	for _, item := range componentItems {
		if err := plugins.Store.DeleteComponent(ctx, item.UID); err != nil {
			return fmt.Errorf("failed to delete Component: %w", err)
		}
	}
	// Delete all componentendpoints
	componentendpointItems, err := storage.Querycomponentendpoints(ctx).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query componentendpoints: %w", err)
	}
	for _, item := range componentendpointItems {
		if err := plugins.Store.DeleteComponentEndpoint(ctx, item.UID); err != nil {
			return fmt.Errorf("failed to delete ComponentEndpoint: %w", err)
		}
	}
	// Delete all ethernetinterfaces
	ethernetinterfaceItems, err := storage.Queryethernetinterfaces(ctx).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query ethernetinterfaces: %w", err)
	}
	for _, item := range ethernetinterfaceItems {
		if err := plugins.Store.DeleteEthernetInterface(ctx, item.UID); err != nil {
			return fmt.Errorf("failed to delete EthernetInterface: %w", err)
		}
	}
	// Delete all groups
	groupItems, err := storage.Querygroups(ctx).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query groups: %w", err)
	}
	for _, item := range groupItems {
		if err := plugins.Store.DeleteGroup(ctx, item.UID); err != nil {
			return fmt.Errorf("failed to delete Group: %w", err)
		}
	}
	// Delete all hardwares
	hardwareItems, err := storage.Queryhardwares(ctx).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query hardwares: %w", err)
	}
	for _, item := range hardwareItems {
		if err := storage.DeleteHardware(ctx, item.UID); err != nil {
			return fmt.Errorf("failed to delete Hardware: %w", err)
		}
	}
	// Delete all redfishendpoints
	redfishendpointItems, err := storage.Queryredfishendpoints(ctx).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query redfishendpoints: %w", err)
	}
	for _, item := range redfishendpointItems {
		if err := plugins.Store.DeleteRedfishEndpoint(ctx, item.UID); err != nil {
			return fmt.Errorf("failed to delete RedfishEndpoint: %w", err)
		}
	}
	// Delete all serviceendpoints
	serviceendpointItems, err := storage.Queryserviceendpoints(ctx).All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query serviceendpoints: %w", err)
	}
	for _, item := range serviceendpointItems {
		if err := plugins.Store.DeleteServiceEndpoint(ctx, item.UID); err != nil {
			return fmt.Errorf("failed to delete ServiceEndpoint: %w", err)
		}
	}
	return nil
}
