package kritiimages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/png"

	"github.com/fogleman/gg"
	"github.com/kritihq/kriti-images/pkg/kritiimages/models"
)

// TODO: fetch from configs
const fontPath = "/usr/share/fonts/arial/ARIAL.TTF"

var (
	ErrSourceTemplateNotFound = errors.New("source template not found")
	ErrTemplateNotFound       = errors.New("failed to get template")
	ErrTemplateRenderFailed   = errors.New("failed to render template")
	ErInvalidTemplateSources  = errors.New("invalid templatesource instance provided")
)

// RenderTemplate renders a template into an image.
//
// Accepts context.Context for cancellation and timeout.
// Accepts template name.
// Accepts map of variables to substitute in the template.
//
// Returns the rendered image as a buffer.
func (k *KritiImages) RenderTemplate(ctx context.Context, templateName string, vars map[string]string) (*bytes.Buffer, error) {
	root, err := k.DefaultTemplateSources.GetTemplateSubstituted(ctx, templateName, vars)
	if err != nil {
		return nil, errors.Join(ErrTemplateRenderFailed, err)
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
	if err := k.renderNode(ctx, dc, root); err != nil {
		return nil, err
	}

	// Out as buffer
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf, nil
}

func (k *KritiImages) renderNode(ctx context.Context, dc *gg.Context, node *models.Node) error {
	switch node.ClassName {
	default: // ignore anything else, focus on child nodes
		for _, child := range node.Children {
			if err := k.renderNode(ctx, dc, &child); err != nil {
				return err
			}
		}
	case "Image":
		return k.renderImageNode(ctx, dc, &node.Attrs)
	case "Text":
		return k.renderTextNode(dc, &node.Attrs)
	}
	return nil
}

func (k *KritiImages) renderImageNode(ctx context.Context, dc *gg.Context, attrs *models.Attrs) error {
	path := attrs.Path
	if path == "" {
		return nil
	}
	source := k.getImageSource(path)
	img, _, err := source.GetImage(ctx, path)
	if err != nil {
		return ErrSourceImageNotFound
	}

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

	dc.Push()
	dc.Translate(x, y)
	dc.Scale(scaleX, scaleY)
	dc.DrawImage(img, 0, 0)
	dc.Pop()
	return nil
}

func (k *KritiImages) renderTextNode(dc *gg.Context, attrs *models.Attrs) error {
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
