package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fogleman/gg"
)

// TODO: fetch from configs
const fontPath = "/usr/share/fonts/arial/ARIAL.TTF"

// Structs for parsing the JSON template
type Node struct {
	ClassName string `json:"className"`
	Attrs     Attrs  `json:"attrs"`
	Children  []Node `json:"children,omitempty"`
}

type Attrs struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	ScaleX float64 `json:"scaleX"`
	ScaleY float64 `json:"scaleY"`

	// text
	FontSize float64 `json:"fontSize"`
	Text     string  `json:"text"`
	Fill     string  `json:"fill"` // hex color code

	// image
	Path string `json:"path"`
}

func substituteVariables(template string, variables map[string]string) string {
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.Trim(match, "{} ")
		if val, ok := variables[key]; ok {
			return val
		}
		return match
	})
}

func RenderTemplate(templateJSON string, variables map[string]string) (*bytes.Buffer, error) {
	// Substitute variables
	substituted := substituteVariables(templateJSON, variables)

	// Parse JSON
	var root Node
	if err := json.Unmarshal([]byte(substituted), &root); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}

	// Canvas setup
	width, height := root.Attrs.Width, root.Attrs.Width
	if width <= 0 || height <= 0 || width > 2048 || height > 2048 {
		return nil, fmt.Errorf("invalid canvas size")
	}
	dc := gg.NewContext(width, height)

	// set base layer color
	fill := root.Attrs.Fill
	if fill == "" {
		fill = "#00000000"
	}
	dc.Push()
	dc.SetHexColor(fill)
	dc.DrawRectangle(0, 0, float64(width), float64(height))
	dc.Fill()
	dc.Pop()

	// Render recursively
	if err := renderNode(dc, root); err != nil {
		return nil, err
	}

	// Out as buffer
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf, nil
}

func renderNode(dc *gg.Context, node Node) error {
	switch node.ClassName {
	default: // ignore anything else, focus on child nodes
		for _, child := range node.Children {
			if err := renderNode(dc, child); err != nil {
				return err
			}
		}
	case "Image":
		return renderImageNode(dc, &node.Attrs)
	case "Text":
		return renderTextNode(dc, &node.Attrs)
	}
	return nil
}

func renderImageNode(dc *gg.Context, attrs *Attrs) error {
	path := attrs.Path
	x := attrs.X
	y := attrs.Y
	scaleX := attrs.ScaleX
	if scaleX == 0 {
		scaleX = 1
	}
	scaleY := attrs.ScaleY
	if scaleY == 0 {
		scaleY = 1
	}

	if path == "" {
		return nil
	}
	imgFile, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to open image %s: %w", path, err)
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("failed to decode image %s: %w", path, err)
	}

	dc.Push()
	dc.Translate(x, y)
	dc.Scale(scaleX, scaleY)
	dc.DrawImage(img, 0, 0)
	dc.Pop()
	return nil
}

func renderTextNode(dc *gg.Context, attrs *Attrs) error {
	text := attrs.Text
	x := attrs.X
	y := attrs.Y
	fontSize := attrs.FontSize
	fill := attrs.Fill
	scaleX := attrs.ScaleX
	if scaleX == 0 {
		scaleX = 1
	}
	scaleY := attrs.ScaleY
	if scaleY == 0 {
		scaleY = 1
	}

	if fontSize == 0 {
		fontSize = 24
	}
	if fill == "" {
		fill = "#000000"
	}

	dc.Push()
	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		return fmt.Errorf("failed to load font: %w", err)
	}
	dc.SetHexColor(fill)
	dc.Scale(scaleX, scaleY)
	dc.DrawStringAnchored(text, x, y, 0, 1.1)
	dc.Pop()
	return nil
}
