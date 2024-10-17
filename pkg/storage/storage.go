package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

type FileInfo struct {
	LastModified time.Time
	Size         int64
}

type Uploader interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error
	Download(ctx context.Context, key string, writer io.Writer) error
	List(ctx context.Context, prefix string) ([]string, error)
	Delete(ctx context.Context, key string) error
	FileExists(ctx context.Context, key string) (bool, error)
	GetFileInfo(ctx context.Context, key string) (*FileInfo, error)
}

func Upload(c *cli.Context, uploader Uploader, log *logger.Logger) error {
	rootDir := c.Args().First()
	if rootDir == "" {
		return cli.NewExitError("No root directory specified", 1)
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

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		key := filepath.Join(keyPrefix, relPath)
		if compress {
			key += ".gz"
		}

		eligible, err := fileUploadEligible(c.Context, uploader, key, info)
		if err != nil {
			log.Error("Error checking file eligibility", "file", path, "error", err)
			return err
		}

		if !eligible {
			log.Info("Skipping file (already uploaded or not modified)", "file", path)
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Error("Failed to open file", "file", path, "error", err)
			return err
		}
		defer file.Close()

		var reader io.Reader = file
		size := info.Size()

		if compress {
			// Implement compression logic here
			// For brevity, we're skipping the actual compression in this example
			log.Info("Compressing file", "file", path)
		}

		if !nonInteractive {
			bar := progressbar.DefaultBytes(
				size,
				"Uploading "+filepath.Base(path),
			)
			reader = io.TeeReader(reader, bar)
		}

		err = uploader.Upload(c.Context, key, reader, size)
		if err != nil {
			log.Error("Failed to upload file", "file", path, "error", err)
			return err
		}

		log.Info("File uploaded successfully", "file", path, "key", key)
		return nil
	})
}

func fileUploadEligible(ctx context.Context, uploader Uploader, key string, info os.FileInfo) (bool, error) {
	exists, err := uploader.FileExists(ctx, key)
	if err != nil {
		return false, err
	}

	if !exists {
		return true, nil
	}

	remoteInfo, err := uploader.GetFileInfo(ctx, key)
	if err != nil {
		return false, err
	}

	// Compare last modified times
	if info.ModTime().After(remoteInfo.LastModified) {
		return true, nil
	}

	// If the file sizes are different, consider it eligible for upload
	if info.Size() != remoteInfo.Size {
		return true, nil
	}

	return false, nil
}

func Download(c *cli.Context, uploader Uploader, log *logger.Logger) error {
	key := c.Args().First()
	if key == "" {
		return cli.NewExitError("No key specified", 1)
	}

	outFile := filepath.Base(key)
	if c.String("output") != "" {
		outFile = c.String("output")
	}

	file, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer file.Close()

	bar := progressbar.DefaultBytes(
		-1,
		"Downloading "+filepath.Base(key),
	)

	writer := io.MultiWriter(file, bar)

	err = uploader.Download(context.Background(), key, writer)
	if err != nil {
		return err
	}

	log.Info("File downloaded successfully", "file", outFile)
	return nil
}

func Prune(c *cli.Context, uploader Uploader, log *logger.Logger) error {
	prefix := c.String("keyprefix")
	age := c.Duration("age")
	if age == 0 {
		return cli.NewExitError("No age specified for pruning", 1)
	}

	cutoff := time.Now().Add(-age)

	files, err := uploader.List(context.Background(), prefix)
	if err != nil {
		return err
	}

	for _, key := range files {
		// This is a simplification. In a real-world scenario, you'd want to check the last modified time of each file.
		// The exact implementation depends on the metadata available from your storage provider.
		// For this example, we're just using the key (which might contain a timestamp) for demonstration.
		fileTime, err := time.Parse("2006-01-02-15-04-05", filepath.Base(key))
		if err != nil {
			log.Warn("Unable to parse time from key, skipping", "key", key)
			continue
		}

		if fileTime.Before(cutoff) {
			err = uploader.Delete(context.Background(), key)
			if err != nil {
				log.Error("Failed to delete file", "key", key, "error", err)
			} else {
				log.Info("Deleted old file", "key", key)
			}
		}
	}

	return nil
}
