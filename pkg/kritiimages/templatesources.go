package kritiimages

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kritihq/kriti-images/internal/templatesources"
	"github.com/kritihq/kriti-images/pkg/kritiimages/models"
)

// TemplateSource represents an source to retrieve templates from.
type TemplateSource interface {
	// GetTemplate retrieves the template with name `fileName` from the source.
	// If the template is present it is returned as `Node`.
	//
	// In case of any error or no template found, `error` is returned and other
	// return values are null and empty.
	GetTemplateSubstituted(ctx context.Context, fileName string, vars map[string]string) (*models.Node, error)

	// NOTE: UploadTemplate is not yet required
}

func NewTemplateSourceLocal(basePath string) *templatesources.TemplateSourceLocal {
	return &templatesources.TemplateSourceLocal{
		BasePath: basePath,
	}
}

func NewTemplateSourceS3(ctx context.Context, bucket string, client *s3.Client) *templatesources.TemplateSourceS3 {
	return &templatesources.TemplateSourceS3{
		Bucket: bucket,
		Client: client,
	}
}
