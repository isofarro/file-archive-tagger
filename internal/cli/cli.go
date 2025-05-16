package cli

import (
	"fmt"
	"go-fart/internal/fileops"
	"os"
	"path/filepath"
	"strings"
)

type CLI struct {
	taxonomyManager TaxonomyManager
	db              DatabaseManager
}

type TaxonomyManager interface {
	InitTaxonomy(name string) error
	TagFile(filePath, taxonomyName, tagValue string) error
	SearchByTag(taxonomyName, tagValue string) ([]string, error)
}

type DatabaseManager interface {
	FileExists(hash string) (bool, error)
	AddFile(filename, path, hash string, size int64, modifiedAt string) error
	GetFilePathByHash(hash string) (string, error)
}

func New(tm TaxonomyManager, db DatabaseManager) *CLI {
	return &CLI{
		taxonomyManager: tm,
		db:              db,
	}
}

// HandleTaxonomyCommand processes taxonomy-related commands
func (c *CLI) HandleTaxonomyCommand(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fart taxonomy <init|tag|search> [arguments]")
	}

	switch args[1] {
	case "init":
		if len(args) != 3 {
			return fmt.Errorf("usage: fart taxonomy init <taxonomy-name>")
		}
		return c.taxonomyManager.InitTaxonomy(args[2])

	default:
		return fmt.Errorf("unknown taxonomy subcommand: %s", args[1])
	}
}

// HandleTagCommand processes tag-related commands
func (c *CLI) HandleTagCommand(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: fart tag <file> <tag-value> [--taxonomy <taxonomy-name>]")
	}

	filePath := args[1]
	tagValue := args[2]
	taxonomyName := "tags" // default taxonomy

	// Check for taxonomy flag
	for i := 3; i < len(args)-1; i++ {
		if strings.HasPrefix(args[i], "--") {
			name, err := parseTaxonomyFlag(args[i])
			if err != nil {
				return err
			}
			taxonomyName = name
			tagValue = args[i+1]
			break
		}
	}

	// Convert to absolute path if necessary
	if !filepath.IsAbs(filePath) {
		var err error
		filePath, err = filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	return c.taxonomyManager.TagFile(filePath, taxonomyName, tagValue)
}

// HandleSearchCommand processes search-related commands
func (c *CLI) HandleSearchCommand(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: fart search --<taxonomy-name> <tag-value>")
	}

	taxonomyFlag := args[1]
	if !strings.HasPrefix(taxonomyFlag, "--") {
		return fmt.Errorf("invalid taxonomy flag format: must start with --")
	}

	taxonomyName := strings.TrimPrefix(taxonomyFlag, "--")
	tagValue := args[2]

	files, err := c.taxonomyManager.SearchByTag(taxonomyName, tagValue)
	if err != nil {
		return err
	}

	// Print results
	for _, file := range files {
		fmt.Println(file)
	}
	return nil
}

// HandleCheckCommand processes check-related commands
func (c *CLI) HandleCheckCommand(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fart check <file-path>")
	}

	filePath := args[1]
	fileInfo, err := fileops.GetFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	matchingPath, err := c.db.GetFilePathByHash(fileInfo.Hash)
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	if matchingPath != "" {
		fmt.Printf("File already exists at: %s\n", matchingPath)
	} else {
		fmt.Printf("File does not exist in the database\n")
	}

	return nil
}

// Helper function to parse taxonomy flags
func parseTaxonomyFlag(flag string) (string, error) {
	if !strings.HasPrefix(flag, "--") {
		return "", fmt.Errorf("invalid taxonomy flag format: must start with --")
	}
	name := strings.TrimPrefix(flag, "--")
	if name == "" {
		return "", fmt.Errorf("taxonomy name cannot be empty")
	}
	return name, nil
}

// HandleAddCommand processes add-related commands
func (c *CLI) HandleAddCommand(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: fart add <file|directory|pattern>")
	}

	for _, pattern := range args[1:] {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}

		// If no matches found and the pattern doesn't contain wildcards,
		// treat it as a direct file/directory path
		if len(matches) == 0 && !strings.ContainsAny(pattern, "*?[]") {
			matches = []string{pattern}
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				fmt.Printf("Warning: skipping %s: %v\n", match, err)
				continue
			}

			if info.IsDir() {
				err = c.addDirectory(match)
			} else {
				err = c.addFile(match)
			}

			if err != nil {
				fmt.Printf("Warning: error processing %s: %v\n", match, err)
			}
		}
	}

	return nil
}

// addFile adds a single file to the database
func (c *CLI) addFile(path string) error {
	fileInfo, err := fileops.GetFileInfo(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Get relative path from current directory
	relPath, err := filepath.Rel(".", filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Add file to database
	err = c.db.AddFile(
		filepath.Base(path),
		relPath,
		fileInfo.Hash,
		fileInfo.Size,
		fileInfo.ModifiedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add file to database: %w", err)
	}

	fmt.Printf("Added %s\n", path)
	return nil
}

// addDirectory recursively adds all files in a directory
func (c *CLI) addDirectory(path string) error {
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(filepath.Base(filePath), ".") {
			return nil
		}

		return c.addFile(filePath)
	})
}
