package kritiimages

import (
	"context"
	"image"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kritihq/kriti-images/internal/imagesources"
)

// ImageSource represents an source to retrieve images from.
type ImageSource interface {
	// GetImage retrieves the image with name `fileName` from the source.
	// If the image is present it is returned as `image.Image` along with its
	// format i.e. extension (JPEG, PNG or WEBP).
	//
	// In case of any error or no image found, `error` is returned and other
	// return values are null and empty.
	GetImage(ctx context.Context, fileName string) (image.Image, string, error)

	// UploadImage uploads the image with name `fileName` to the source.
	// If the image is present it is returned as `image.Image` along with its
	// format i.e. extension (JPEG, PNG or WEBP).
	//
	// In case of any error or no image found, `error` is returned and other
	// return values are null and empty.
	//
	// Not all ImageSources support upload, e.g. URL based ImageSource which pulls images from any HTTP(s) URL.
	//
	// NOTE: method is experimental and may be removed in future.
	UploadImage(ctx context.Context, fileName string, file image.Image) error
}

func NewImageSourceLocal(basePath string, validations *imagesources.SourceImageValidations) *imagesources.ImageSourceLocal {
	return &imagesources.ImageSourceLocal{
		BasePath:               basePath,
		SourceImageValidations: *validations,
	}
}

func NewImageSourceURL(validations *imagesources.SourceImageValidations) *imagesources.ImageSourceHTTP {
	return &imagesources.ImageSourceHTTP{
		SourceImageValidations: *validations,
	}
}

func NewImageSourceS3(ctx context.Context, bucket string, client *s3.Client, validations *imagesources.SourceImageValidations) *imagesources.ImageSourceS3 {
	return &imagesources.ImageSourceS3{
		SourceImageValidations: *validations,
		Bucket:                 bucket,
		Client:                 client,
	}
}
