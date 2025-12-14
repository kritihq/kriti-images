package imagesources

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"strings"
)

type ImageSourceHTTP struct {
	SourceImageValidations
}

func (i ImageSourceHTTP) GetImage(ctx context.Context, url string) (image.Image, string, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, "", fmt.Errorf("invalid URL")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	n, err := io.Copy(buf, resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	if err := validateImageSize(n, i.MaxFileSizeInBytes); err != nil {
		return nil, "", err
	}

	img, format, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	if err := validateImageDimensions(img.Bounds().Dx(), img.Bounds().Dy(), i.MaxImageDimension); err != nil {
		return nil, "", err
	}

	return img, format, nil
}

// UploadImage is not supported for URL source
func (i ImageSourceHTTP) UploadImage(ctx context.Context, fileName string, file image.Image) error {
	return fmt.Errorf("upload not supported for HTTP source")
}
