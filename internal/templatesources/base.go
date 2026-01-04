// package templatesources handles various sources from where an template can be retrieved.
// e.g. local disk, AWS S3 or Cloudflare R2
package templatesources

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kritihq/kriti-images/pkg/kritiimages/models"
)

// TODO: add other S3 compatible sources

// TemplateSourceLocal represents the machine's local disk as an template source.
type TemplateSourceLocal struct {
	BasePath string // base path of the mounted disk
}

func (i *TemplateSourceLocal) GetTemplateSubstituted(ctx context.Context, fileName string, vars map[string]string) (*models.Node, error) {
	// Ensure the path is safe and doesn't contain directory traversal
	cleanPath := filepath.Clean(fileName)
	if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid template path")
	}

	fullPath := filepath.Join(i.BasePath, cleanPath)
	_, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat template: %w", err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("template file not found")
	}

	return substituteVarsAndUnmarshalToRootNode(data, vars)
}

func substituteVarsAndUnmarshalToRootNode(data []byte, vars map[string]string) (*models.Node, error) {
	tmpl, err := template.New("templateJSON").Parse(string(data))
	if err != nil {
		return nil, errors.New("invalid JSON in template file")
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, vars); err != nil {
		return nil, errors.New("failed to execute template")
	}

	return unmarshalToRootNode([]byte(buf.Bytes()))
}

func unmarshalToRootNode(data []byte) (*models.Node, error) {
	var root models.Node
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to deserialize template: %w", err)
	}

	return &root, nil
}
