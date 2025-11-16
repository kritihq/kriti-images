package imagesources

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chai2010/webp"
)

type ImageSourceS3 struct {
	SourceImageValidations
	Bucket string
	Client *s3.Client
}

func NewImageSourceS3(ctx context.Context, bucket string, validations *SourceImageValidations) (*ImageSourceS3, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return &ImageSourceS3{
		SourceImageValidations: *validations,
		Bucket:                 bucket,
		Client:                 client,
	}, nil
}

func (i *ImageSourceS3) GetImage(ctx context.Context, fileName string) (image.Image, string, error) {
	cleanPath := filepath.Clean(fileName)
	if strings.Contains(cleanPath, "..") {
		return nil, "", fmt.Errorf("invalid image path")
	}

	resp, err := i.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(i.Bucket),
		Key:    aws.String(cleanPath),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image from S3: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	if err := validateImageSize(int64(buf.Len()), i.MaxFileSizeInBytes); err != nil {
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

func (i *ImageSourceS3) UploadImage(ctx context.Context, fileName string, file image.Image) error {
	cleanPath := filepath.Clean(fileName)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid image path")
	}

	if err := validateImageDimensions(file.Bounds().Dx(), file.Bounds().Dy(), i.MaxImageDimension); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(buf, file, &jpeg.Options{Quality: 85}); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	case ".png":
		if err := png.Encode(buf, file); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	case ".webp":
		if err := webp.Encode(buf, file, &webp.Options{Quality: 85}); err != nil {
			return fmt.Errorf("failed to encode WebP: %w", err)
		}
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	if err := validateImageSize(int64(buf.Len()), i.MaxFileSizeInBytes); err != nil {
		return err
	}

	_, err := i.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(i.Bucket),
		Key:    aws.String(cleanPath),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to S3: %w", err)
	}

	return nil
}
