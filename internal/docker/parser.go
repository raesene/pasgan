package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/raesene/pasgan/pkg/utils"
)

// ManifestItem represents an item in the manifest.json file
type ManifestItem struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// ImageMetadata represents Docker image metadata
type ImageMetadata struct {
	ID           string              `json:"id,omitempty"`
	Config       Config              `json:"config"`
	RepoTags     []string            `json:"RepoTags"`
	Architecture string              `json:"architecture"`
	OS           string              `json:"os"`
	Created      time.Time           `json:"created"`
	DockerVersion string             `json:"docker_version"`
	History      []History           `json:"history"`
	RootFS       RootFS              `json:"rootfs"`
	Layers       []string            `json:"layers"`
	LayerConfigs map[string]*LayerConfig `json:"-"`
}

// RootFS represents the rootfs configuration
type RootFS struct {
	Type    string   `json:"type"`
	DiffIDs []string `json:"diff_ids"`
}

// Config represents Docker image config
type Config struct {
	Hostname     string              `json:"Hostname"`
	Domainname   string              `json:"Domainname"`
	User         string              `json:"User"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts"`
	Env          []string            `json:"Env"`
	Cmd          []string            `json:"Cmd"`
	WorkingDir   string              `json:"WorkingDir"`
	Entrypoint   []string            `json:"Entrypoint"`
	Labels       map[string]string   `json:"Labels"`
	StopSignal   string              `json:"StopSignal,omitempty"`
	Volumes      map[string]struct{} `json:"Volumes,omitempty"`
}

// History represents a layer history entry
type History struct {
	Created    string `json:"created"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

// LayerConfig represents the configuration for a specific layer
type LayerConfig struct {
	ID      string    `json:"id"`
	Created time.Time `json:"created"`
	Parent  string    `json:"parent,omitempty"`
	Config  Config    `json:"config,omitempty"`
}

// Parser provides functionality to parse Docker images
type Parser struct {
	imagePath string
	workDir   string
}

// NewParser creates a new Docker image parser
func NewParser(imagePath string) (*Parser, error) {
	// Create a temporary directory for extraction
	workDir, err := os.MkdirTemp("", "pasgan-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &Parser{
		imagePath: imagePath,
		workDir:   workDir,
	}, nil
}

// Parse extracts and analyzes a Docker image
func (p *Parser) Parse() (*ImageMetadata, error) {
	// Extract the tar file
	if err := utils.ExtractTar(p.imagePath, p.workDir); err != nil {
		return nil, fmt.Errorf("failed to extract image archive: %w", err)
	}

	// Read manifest.json
	manifestPath := filepath.Join(p.workDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest.json: %w", err)
	}

	var manifest []ManifestItem
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
	}

	if len(manifest) == 0 {
		return nil, fmt.Errorf("manifest.json contains no images")
	}

	// Get the first image from the manifest
	item := manifest[0]

	// Read the image config
	configPath := filepath.Join(p.workDir, item.Config)
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var imageMetadata ImageMetadata
	if err := json.Unmarshal(configData, &imageMetadata); err != nil {
		return nil, fmt.Errorf("failed to parse image config: %w", err)
	}

	// Set layer paths and repo tags
	imageMetadata.Layers = item.Layers
	imageMetadata.RepoTags = item.RepoTags

	// Parse layer configs
	imageMetadata.LayerConfigs = make(map[string]*LayerConfig)
	for _, layerPath := range item.Layers {
		layerDir := filepath.Dir(layerPath)
		jsonPath := filepath.Join(p.workDir, filepath.Join(layerDir, "json"))
		
		// Skip if json file does not exist
		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			continue
		}

		jsonData, err := os.ReadFile(jsonPath)
		if err != nil {
			// Skip if can't read
			continue
		}

		var layerConfig LayerConfig
		if err := json.Unmarshal(jsonData, &layerConfig); err != nil {
			// Skip if can't parse
			continue
		}

		id := filepath.Base(layerDir)
		imageMetadata.LayerConfigs[id] = &layerConfig
	}

	return &imageMetadata, nil
}

// Cleanup removes temporary files
func (p *Parser) Cleanup() error {
	if p.workDir != "" {
		return os.RemoveAll(p.workDir)
	}
	return nil
}