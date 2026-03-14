package templatesources

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kritihq/kriti-images/pkg/kritiimages/models"
)

type TemplateSourceS3 struct {
	Bucket string
	Client *s3.Client
}

func (i *TemplateSourceS3) GetTemplateSubstituted(ctx context.Context, fileName string, vars map[string]string) (*models.Node, error) {
	cleanPath := filepath.Clean(fileName)
	if strings.Contains(cleanPath, "..") {
		return nil, errors.New("invalid template path")
	}

	resp, err := i.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(i.Bucket),
		Key:    aws.String(cleanPath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get template from S3: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return substituteVarsAndUnmarshalToRootNode([]byte(buf.Bytes()), vars)
}
