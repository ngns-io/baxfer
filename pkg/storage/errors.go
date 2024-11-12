package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// UserError represents a user-friendly error message
type UserError struct {
	Message string
	Cause   error
}

func (e *UserError) Error() string {
	return e.Message
}

func (e *UserError) Unwrap() error {
	return e.Cause
}

// formatDownloadError converts provider-specific errors into user-friendly messages
func formatDownloadError(provider, key string, err error) error {
	if err == nil {
		return nil
	}

	// Common error handling for S3-compatible services (S3, R2, B2S3)
	var (
		nsk      *types.NoSuchKey
		notFound *types.NotFound
	)
	if errors.As(err, &nsk) || errors.As(err, &notFound) ||
		strings.Contains(err.Error(), "NoSuchKey") ||
		strings.Contains(err.Error(), "status code: 404") ||
		strings.Contains(err.Error(), "StatusCode: 404") {
		return &UserError{
			Message: fmt.Sprintf("File not found: %s", key),
			Cause:   err,
		}
	}

	// Check for access denied errors
	if strings.Contains(strings.ToLower(err.Error()), "access denied") ||
		strings.Contains(err.Error(), "status code: 403") {
		return &UserError{
			Message: fmt.Sprintf("Access denied to file: %s. Please check your credentials and permissions.", key),
			Cause:   err,
		}
	}

	// Provider-specific error messages
	switch provider {
	case "s3":
		return formatS3Error(key, err)
	case "r2":
		return formatR2Error(key, err)
	case "b2":
		return formatB2Error(key, err)
	case "b2s3":
		return formatB2S3Error(key, err)
	case "sftp":
		return formatSFTPError(key, err)
	default:
		return &UserError{
			Message: fmt.Sprintf("Failed to download file: %s", key),
			Cause:   err,
		}
	}
}

func formatS3Error(key string, err error) error {
	// Add specific AWS S3 error handling
	if strings.Contains(err.Error(), "InvalidAccessKeyId") {
		return &UserError{
			Message: "Invalid AWS credentials. Please check your access key ID.",
			Cause:   err,
		}
	}
	if strings.Contains(err.Error(), "SignatureDoesNotMatch") {
		return &UserError{
			Message: "Invalid AWS credentials. Please check your secret access key.",
			Cause:   err,
		}
	}
	return &UserError{
		Message: fmt.Sprintf("Error downloading from S3: %s", key),
		Cause:   err,
	}
}

func formatR2Error(key string, err error) error {
	// Add specific R2 error handling
	if strings.Contains(err.Error(), "InvalidAccessKeyId") {
		return &UserError{
			Message: "Invalid Cloudflare R2 credentials. Please check your access key ID.",
			Cause:   err,
		}
	}
	return &UserError{
		Message: fmt.Sprintf("Error downloading from R2: %s", key),
		Cause:   err,
	}
}

func formatB2Error(key string, err error) error {
	if strings.Contains(err.Error(), "401") {
		return &UserError{
			Message: "Invalid B2 credentials. Please check your application key and key ID.",
			Cause:   err,
		}
	}
	return &UserError{
		Message: fmt.Sprintf("Error downloading from B2: %s", key),
		Cause:   err,
	}
}

func formatB2S3Error(key string, err error) error {
	if strings.Contains(err.Error(), "InvalidAccessKeyId") {
		return &UserError{
			Message: "Invalid B2 credentials. Please check your access key ID.",
			Cause:   err,
		}
	}
	return &UserError{
		Message: fmt.Sprintf("Error downloading from B2 S3: %s", key),
		Cause:   err,
	}
}

func formatSFTPError(key string, err error) error {
	if strings.Contains(err.Error(), "permission denied") {
		return &UserError{
			Message: "Permission denied. Please check your SFTP credentials and permissions.",
			Cause:   err,
		}
	}
	if strings.Contains(err.Error(), "connection refused") {
		return &UserError{
			Message: "Could not connect to SFTP server. Please check your connection settings.",
			Cause:   err,
		}
	}
	return &UserError{
		Message: fmt.Sprintf("Error downloading via SFTP: %s", key),
		Cause:   err,
	}
}
