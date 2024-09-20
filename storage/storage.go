package storage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"path/filepath"
)

type S3 struct {
	Client *minio.Client
	Bucket string
}

func NewS3(bucket string, region string, endpoint string, accessKey string, secretKey string) *S3 {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
		Region: region,
	})
	if err != nil {
		return nil
	}
	return &S3{Client: minioClient, Bucket: bucket}
}

func (storage *S3) Get(ctx context.Context, key string) ([]byte, error) {
	object, err := storage.Client.GetObject(ctx, storage.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (storage *S3) Add(ctx context.Context, key string, data []byte) error {
	reader := bytes.NewReader(data)
	_, err := storage.Client.PutObject(ctx, storage.Bucket, key, reader, int64(reader.Len()), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})

	if err != nil {
		return err
	}
	return nil
}

func (storage *S3) Delete(ctx context.Context, key string) error {
	err := storage.Client.RemoveObject(ctx, storage.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (storage *S3) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var objects []string
	objectCh := storage.Client.ListObjects(ctx, storage.Bucket, minio.ListObjectsOptions{
		Prefix: prefix,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		fileName := filepath.Base(object.Key)
		objects = append(objects, fileName)
	}
	return objects, nil
}
