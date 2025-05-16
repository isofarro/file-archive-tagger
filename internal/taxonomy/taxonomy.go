package taxonomy

import (
    "fmt"
    "strings"
)

// Manager handles taxonomy-related operations
type Manager struct {
    db TaxonomyDB
}

// TaxonomyDB interface defines the required database operations
type TaxonomyDB interface {
    AddTaxonomy(name string) error
    TagFile(filePath, taxonomyName, tagName string) error
    SearchByTag(taxonomyName, tagName string) ([]string, error)
}

// New creates a new taxonomy manager
func New(db TaxonomyDB) *Manager {
    return &Manager{db: db}
}

// InitTaxonomy creates a new taxonomy
func (m *Manager) InitTaxonomy(name string) error {
    if name == "" {
        return fmt.Errorf("taxonomy name cannot be empty")
    }
    name = strings.ToLower(strings.TrimSpace(name))
    return m.db.AddTaxonomy(name)
}

// TagFile adds a tag to a file under a specific taxonomy
func (m *Manager) TagFile(filePath, taxonomyName, tagValue string) error {
    if filePath == "" || taxonomyName == "" || tagValue == "" {
        return fmt.Errorf("file path, taxonomy name, and tag value are required")
    }
    
    taxonomyName = strings.ToLower(strings.TrimSpace(taxonomyName))
    return m.db.TagFile(filePath, taxonomyName, tagValue)
}

// SearchByTag searches for files with a specific tag under a taxonomy
func (m *Manager) SearchByTag(taxonomyName, tagValue string) ([]string, error) {
    if taxonomyName == "" || tagValue == "" {
        return nil, fmt.Errorf("taxonomy name and tag value are required")
    }
    
    taxonomyName = strings.ToLower(strings.TrimSpace(taxonomyName))
    return m.db.SearchByTag(taxonomyName, tagValue)
}

// ParseTaxonomyFlag parses a taxonomy flag in the format --taxonomyName
func ParseTaxonomyFlag(flag string) (string, error) {
    if !strings.HasPrefix(flag, "--") {
        return "", fmt.Errorf("invalid taxonomy flag format: must start with --")
    }
    name := strings.TrimPrefix(flag, "--")
    if name == "" {
        return "", fmt.Errorf("taxonomy name cannot be empty")
    }
    return name, nil
}