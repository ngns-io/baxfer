package storage

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

type Uploader interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error
	Download(ctx context.Context, key string, writer io.Writer) error
	List(ctx context.Context, prefix string) ([]string, error)
	Delete(ctx context.Context, key string) error
	FileExists(ctx context.Context, key string) (bool, error)
	GetFileInfo(ctx context.Context, key string) (*FileInfo, error)
}

func constructKey(rootDir, keyPrefix, path string) (string, error) {
	// Get the relative path
	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return "", err
	}

	// Remove volume name if present (for Windows compatibility)
	relPath = relPath[len(filepath.VolumeName(relPath)):]

	// Convert to forward slashes
	relPath = filepath.ToSlash(relPath)

	// Combine with key prefix
	key := strings.TrimPrefix(filepath.ToSlash(filepath.Join(keyPrefix, relPath)), "/")

	return key, nil
}

func compressFile(sourcePath string, log logger.Logger) (string, int64, error) {
	zipPath := sourcePath + ".zip"
	log.Info("Compressing file", "source", sourcePath, "destination", zipPath)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", 0, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", 0, err
	}
	defer sourceFile.Close()

	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return "", 0, err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return "", 0, err
	}

	// Ensure Windows compatibility by using forward slashes
	header.Name = strings.ReplaceAll(filepath.Base(sourcePath), "\\", "/")
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return "", 0, err
	}

	_, err = io.Copy(writer, sourceFile)
	if err != nil {
		return "", 0, err
	}

	err = zipWriter.Close()
	if err != nil {
		return "", 0, err
	}

	compressedInfo, err := os.Stat(zipPath)
	if err != nil {
		return "", 0, err
	}

	log.Info("File compressed successfully", "file", zipPath, "size", compressedInfo.Size())
	return zipPath, compressedInfo.Size(), nil
}

func Upload(c *cli.Context, uploader Uploader, log logger.Logger) error {
	rootDir := c.Args().First()
	if rootDir == "" {
		return cli.Exit("No root directory specified", 1)
	}

	compress := c.Bool("compress")
	keyPrefix := c.String("keyprefix")
	backupExt := c.String("backupext")
	nonInteractive := c.Bool("non-interactive")

	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != backupExt {
			return nil
		}

		originalKey, err := constructKey(rootDir, keyPrefix, path)
		if err != nil {
			log.Error("Error constructing key", "path", path, "error", err)
			return err
		}

		compressedKey := strings.TrimSuffix(originalKey, filepath.Ext(originalKey)) + ".zip"

		var uploadKey string
		var uploadPath string
		var uploadSize int64

		if compress {
			uploadKey = compressedKey
			eligible, err := fileUploadEligible(c.Context, uploader, uploadKey, info, log)
			if err != nil {
				log.Error("Error checking file eligibility", "file", path, "error", err)
				return err
			}

			if !eligible {
				log.Info("Skipping file (compressed version already uploaded or not modified)", "file", path)
				return nil
			}

			compressedPath, compressedSize, err := compressFile(path, log)
			if err != nil {
				log.Error("Failed to compress file", "file", path, "error", err)
				return err
			}
			uploadPath = compressedPath
			uploadSize = compressedSize
			defer os.Remove(compressedPath) // Clean up the compressed file after upload
		} else {
			uploadKey = originalKey
			eligible, err := fileUploadEligible(c.Context, uploader, uploadKey, info, log)
			if err != nil {
				log.Error("Error checking file eligibility", "file", path, "error", err)
				return err
			}

			if !eligible {
				log.Info("Skipping file (already uploaded or not modified)", "file", path)
				return nil
			}

			uploadPath = path
			uploadSize = info.Size()
		}

		file, err := os.Open(uploadPath)
		if err != nil {
			log.Error("Failed to open file", "file", uploadPath, "error", err)
			return err
		}
		defer file.Close()

		var reader io.Reader = file

		if !nonInteractive {
			bar := progressbar.DefaultBytes(
				uploadSize,
				"Uploading "+filepath.Base(uploadPath),
			)
			reader = io.TeeReader(reader, bar)
		}

		err = uploader.Upload(c.Context, uploadKey, reader, uploadSize)
		if err != nil {
			log.Error("Failed to upload file", "file", uploadPath, "error", err)
			return err
		}

		log.Info("File uploaded successfully", "file", path, "key", uploadKey)
		return nil
	})
}

func Download(c *cli.Context, uploader Uploader, log logger.Logger) error {
	key := c.Args().First()
	if key == "" {
		return cli.Exit("No key specified", 1)
	}

	outFile := filepath.Base(key)
	if c.String("output") != "" {
		outFile = c.String("output")
	}

	file, err := os.Create(outFile)
	if err != nil {
		log.Error("Failed to create output file", "file", outFile, "error", err)
		return err
	}
	defer file.Close()

	nonInteractive := c.Bool("non-interactive")
	var writer io.Writer = file

	if !nonInteractive {
		bar := progressbar.DefaultBytes(
			-1,
			"Downloading "+filepath.Base(key),
		)
		writer = io.MultiWriter(file, bar)
	}

	err = uploader.Download(c.Context, key, writer)
	if err != nil {
		log.Error("Failed to download file", "key", key, "error", err)
		return err
	}

	log.Info("File downloaded successfully", "file", outFile)
	return nil
}

func Prune(c *cli.Context, uploader Uploader, log logger.Logger) error {
	prefix := c.String("keyprefix")
	age := c.Duration("age")
	if age == 0 {
		return cli.Exit("No age specified for pruning", 1)
	}

	cutoff := time.Now().Add(-age)

	files, err := uploader.List(c.Context, prefix)
	if err != nil {
		log.Error("Failed to list files", "error", err)
		return err
	}

	for _, key := range files {
		info, err := uploader.GetFileInfo(c.Context, key)
		if err != nil {
			log.Error("Failed to get file info", "key", key, "error", err)
			continue
		}

		if info.LastModified.Before(cutoff) {
			err = uploader.Delete(c.Context, key)
			if err != nil {
				log.Error("Failed to delete file", "key", key, "error", err)
			} else {
				log.Info("Deleted old file", "key", key)
			}
		}
	}

	return nil
}

func fileUploadEligible(ctx context.Context, uploader Uploader, key string, info os.FileInfo, log logger.Logger) (bool, error) {
	exists, err := uploader.FileExists(ctx, key)
	if err != nil {
		log.Error("Error checking if file exists", "key", key, "error", err)
		return false, err
	}

	if !exists {
		return true, nil
	}

	remoteInfo, err := uploader.GetFileInfo(ctx, key)
	if err != nil {
		log.Error("Error getting remote file info", "key", key, "error", err)
		return false, err
	}

	if info.ModTime().After(remoteInfo.LastModified) {
		log.Info("Local file is newer", "key", key)
		return true, nil
	}

	if info.Size() != remoteInfo.Size {
		log.Info("File sizes differ", "key", key, "local_size", info.Size(), "remote_size", remoteInfo.Size)
		return true, nil
	}

	return false, nil
}
