package s3

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"gosilite/application/infra/imagex"
	"gosilite/application/infra/storage"
	"image"
	"path/filepath"
	"time"
)

func NewStorageImagex(ctx context.Context, ak, sk, bucketName, endpoint, region, domain string) (imagex.ImageX, error) {
	t, err := getS3Client(ctx, ak, sk, bucketName, endpoint, region)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *s3Client) GetUploadAuth(ctx context.Context, opt ...imagex.UploadAuthOpt) (*imagex.SecurityToken, error) {
	return s.GetUploadAuthWithExpire(ctx, time.Hour, opt...)
}

func (s *s3Client) GetUploadAuthWithExpire(ctx context.Context, expire time.Duration, opt ...imagex.UploadAuthOpt) (*imagex.SecurityToken, error) {
	// For S3, we can return a dummy token since we'll use the storage directly
	// In a real implementation, you might want to generate temporary credentials
	return &imagex.SecurityToken{
		AccessKeyID:     "", // Not used with direct storage access
		SecretAccessKey: "", // Not used with direct storage access
		SessionToken:    "", // Not used with direct storage access
		ExpiredTime:     fmt.Sprintf("%d", time.Now().Add(expire).Unix()),
		CurrentTime:     fmt.Sprintf("%d", time.Now().Unix()),
		HostScheme:      "https",
	}, nil
}

func (s *s3Client) GetResourceURL(ctx context.Context, uri string, opts ...imagex.GetResourceOpt) (*imagex.ResourceURL, error) {
	url, err := s.GetObjectUrl(ctx, uri)
	if err != nil {
		return nil, err
	}
	return &imagex.ResourceURL{
		URL: url,
	}, nil
}

func (s *s3Client) Upload(ctx context.Context, data []byte, opts ...imagex.UploadAuthOpt) (*imagex.UploadResult, error) {
	option := &imagex.UploadAuthOption{}
	for _, o := range opts {
		o(option)
	}

	// Generate object key if not provided
	objectKey := ""
	if option.StoreKey != nil {
		objectKey = *option.StoreKey
	} else {
		// Generate a unique key using timestamp and MD5
		md5Hash := md5.Sum(data)
		objectKey = fmt.Sprintf("images/%d/%s", time.Now().Unix(), hex.EncodeToString(md5Hash[:]))
		// Add file extension based on content type
		contentType := detectContentType(data)
		ext := ""
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		}
		objectKey += ext
	}

	// Upload to S3
	contentType := detectContentType(data)
	err := s.PutObject(ctx, objectKey, data, storage.WithContentType(contentType))
	if err != nil {
		return nil, fmt.Errorf("upload to s3 failed: %v", err)
	}

	// Get image info
	imgInfo, err := getImageInfo(data)
	if err != nil {
		// If we can't get image info, continue with minimal info
		imgInfo = &imageInfo{
			width:  0,
			height: 0,
			format: filepath.Ext(objectKey)[1:],
		}
	}

	// Calculate MD5
	md5Hash := md5.Sum(data)
	md5Str := hex.EncodeToString(md5Hash[:])

	// Build response
	result := &imagex.UploadResult{
		RequestId: fmt.Sprintf("%d", time.Now().UnixNano()),
		Result: &imagex.Result{
			Uri:       fmt.Sprintf("/%s", objectKey),
			UriStatus: 2000, // Success
		},
		FileInfo: &imagex.FileInfo{
			Name:        filepath.Base(objectKey),
			Uri:         fmt.Sprintf("/%s", objectKey),
			ImageWidth:  imgInfo.width,
			ImageHeight: imgInfo.height,
			Md5:         md5Str,
			ImageFormat: imgInfo.format,
			ImageSize:   len(data),
			FrameCnt:    1,
			Duration:    0,
		},
	}

	return result, nil
}

func (s *s3Client) GetServerID() string {
	// Not applicable for S3, return empty
	return ""
}

func (s *s3Client) GetUploadHost(ctx context.Context) string {
	// currentHost, ok := ctxcache.Get[string](ctx, consts.HostKeyInCtx)
	// if !ok {
	// 	return ""
	// }
	// return currentHost + consts.ApplyUploadActionURI
	//TODO 待完善
	return ""
}

// Helper functions
type imageInfo struct {
	width  int
	height int
	format string
}

func getImageInfo(data []byte) (*imageInfo, error) {
	img, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &imageInfo{
		width:  img.Width,
		height: img.Height,
		format: format,
	}, nil
}

func detectContentType(data []byte) string {
	// Simple content type detection
	contentType := "application/octet-stream"

	// Check file magic numbers
	if len(data) > 1 {
		if data[0] == 0xFF && data[1] == 0xD8 {
			contentType = "image/jpeg"
		} else if len(data) > 3 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			contentType = "image/png"
		} else if len(data) > 2 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
			contentType = "image/gif"
		}
	}

	return contentType
}
