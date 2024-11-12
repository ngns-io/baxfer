package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPUploader struct {
	client    *sftp.Client
	sshClient *ssh.Client
	basePath  string
	log       logger.Logger
}

func NewSFTPUploader(host string, port int, username, basePath string, log logger.Logger) (*SFTPUploader, error) {
	// Get authentication method from environment
	var authMethod ssh.AuthMethod

	privateKeyPath := os.Getenv("SFTP_PRIVATE_KEY")
	password := os.Getenv("SFTP_PASSWORD")

	if privateKeyPath != "" {
		key, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}

		authMethod = ssh.PublicKeys(signer)
	} else if password != "" {
		authMethod = ssh.Password(password)
	} else {
		return nil, fmt.Errorf("no authentication method provided: set either SFTP_PRIVATE_KEY or SFTP_PASSWORD")
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Consider implementing proper host key verification
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %v", err)
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %v", err)
	}

	// Create base directory if it doesn't exist
	if err := sftpClient.MkdirAll(basePath); err != nil {
		sftpClient.Close()
		sshClient.Close()
		return nil, fmt.Errorf("failed to create base directory: %v", err)
	}

	uploader := &SFTPUploader{
		client:    sftpClient,
		sshClient: sshClient,
		basePath:  basePath,
		log:       log,
	}

	log.Info("Initialized storage provider",
		"provider", "SFTP",
		"host", host,
		"port", port,
		"username", username,
		"basePath", basePath)

	return uploader, nil
}

func (u *SFTPUploader) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	fullPath := filepath.Join(u.basePath, key)
	dir := filepath.Dir(fullPath)

	// Ensure directory exists
	if err := u.client.MkdirAll(dir); err != nil {
		return fmt.Errorf("failed to create directory structure: %v", err)
	}

	// Create remote file
	dstFile, err := u.client.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer dstFile.Close()

	// Copy data
	_, err = io.Copy(dstFile, reader)
	return err
}

func (u *SFTPUploader) Download(ctx context.Context, key string, writer io.Writer) error {
	fullPath := filepath.Join(u.basePath, key)

	srcFile, err := u.client.Open(fullPath)
	if err != nil {
		// Log the original error for debugging
		u.log.Error("Failed to open remote file",
			"path", fullPath,
			"error", err)

		// Handle specific SFTP error cases
		if os.IsNotExist(err) {
			return &UserError{
				Message: fmt.Sprintf("File not found: %s", key),
				Cause:   err,
			}
		}
		if os.IsPermission(err) {
			return &UserError{
				Message: fmt.Sprintf("Permission denied accessing file: %s", key),
				Cause:   err,
			}
		}
		return formatSFTPError(key, err)
	}
	defer srcFile.Close()

	_, err = io.Copy(writer, srcFile)
	if err != nil {
		u.log.Error("Failed to copy file content",
			"path", fullPath,
			"error", err)

		return &UserError{
			Message: fmt.Sprintf("Error reading file content: %s", key),
			Cause:   err,
		}
	}

	return nil
}

func (u *SFTPUploader) List(ctx context.Context, prefix string) ([]string, error) {
	searchPath := filepath.Join(u.basePath, prefix)
	var keys []string

	walker := u.client.Walk(searchPath)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return nil, fmt.Errorf("error walking directory: %v", err)
		}

		path := walker.Path()
		if walker.Stat().IsDir() {
			continue
		}

		// Convert path to key by removing base path
		key := strings.TrimPrefix(path, u.basePath)
		key = strings.TrimPrefix(key, "/")
		keys = append(keys, key)
	}

	return keys, nil
}

func (u *SFTPUploader) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(u.basePath, key)
	return u.client.Remove(fullPath)
}

func (u *SFTPUploader) FileExists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(u.basePath, key)
	_, err := u.client.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u *SFTPUploader) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	fullPath := filepath.Join(u.basePath, key)
	stat, err := u.client.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		LastModified: stat.ModTime(),
		Size:         stat.Size(),
	}, nil
}

func (u *SFTPUploader) Close() error {
	if err := u.client.Close(); err != nil {
		u.sshClient.Close()
		return err
	}
	return u.sshClient.Close()
}
