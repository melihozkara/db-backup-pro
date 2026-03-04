package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Provider implements StorageProvider for S3/MinIO storage
type S3Provider struct {
	storageConfig S3Config
}

// NewS3Provider creates a new S3 storage provider
func NewS3Provider(cfg S3Config) *S3Provider {
	return &S3Provider{
		storageConfig: cfg,
	}
}

// GetType returns the storage type
func (p *S3Provider) GetType() string {
	return "s3"
}

// getClient creates an S3 client
func (p *S3Provider) getClient(ctx context.Context) (*s3.Client, error) {
	// Build custom endpoint resolver for MinIO or custom S3-compatible storage
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if p.storageConfig.Endpoint != "" {
			return aws.Endpoint{
				URL:               p.storageConfig.Endpoint,
				SigningRegion:     p.storageConfig.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(p.storageConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			p.storageConfig.AccessKeyID,
			p.storageConfig.SecretAccessKey,
			"",
		)),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for MinIO
	})

	return client, nil
}

// TestConnection tests the S3 connection
func (p *S3Provider) TestConnection() error {
	ctx := context.Background()
	client, err := p.getClient(ctx)
	if err != nil {
		return err
	}

	// Try to list objects (just to check connection)
	_, err = client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(p.storageConfig.Bucket),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("failed to access bucket: %w", err)
	}

	return nil
}

// Upload uploads a file to S3
func (p *S3Provider) Upload(localPath string, remotePath string) error {
	ctx := context.Background()
	client, err := p.getClient(ctx)
	if err != nil {
		return err
	}

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Build S3 key
	key := path.Join(strings.TrimPrefix(p.storageConfig.Path, "/"), remotePath)

	// Upload file
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.storageConfig.Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// Download downloads a file from S3
func (p *S3Provider) Download(remotePath string, localPath string) error {
	ctx := context.Background()
	client, err := p.getClient(ctx)
	if err != nil {
		return err
	}

	// Build S3 key
	key := path.Join(strings.TrimPrefix(p.storageConfig.Path, "/"), remotePath)

	// Get object
	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.storageConfig.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, result.Body); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// Delete removes a file from S3
func (p *S3Provider) Delete(remotePath string) error {
	ctx := context.Background()
	client, err := p.getClient(ctx)
	if err != nil {
		return err
	}

	// Build S3 key
	key := path.Join(strings.TrimPrefix(p.storageConfig.Path, "/"), remotePath)

	// Delete object
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.storageConfig.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// List returns files in the S3 bucket with the given prefix (paginated)
func (p *S3Provider) List(prefix string) ([]StorageFile, error) {
	ctx := context.Background()
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, err
	}

	// Build S3 prefix
	s3Prefix := path.Join(strings.TrimPrefix(p.storageConfig.Path, "/"), prefix)

	// List objects with pagination
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(p.storageConfig.Bucket),
		Prefix: aws.String(s3Prefix),
	})

	var files []StorageFile
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			// Get relative path
			relPath := strings.TrimPrefix(*obj.Key, s3Prefix)
			relPath = strings.TrimPrefix(relPath, "/")

			if relPath == "" {
				continue
			}

			files = append(files, StorageFile{
				Name:       path.Base(*obj.Key),
				Path:       relPath,
				Size:       *obj.Size,
				ModifiedAt: *obj.LastModified,
				IsDir:      strings.HasSuffix(*obj.Key, "/"),
			})
		}
	}

	return files, nil
}
