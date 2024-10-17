package cli

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/ngns-io/baxfer/pkg/storage"
)

func NewApp(log *logger.Logger) *cli.App {
	app := &cli.App{
		Name:      "baxfer",
		Usage:     "CLI to help manage storage for database backups",
		Version:   "1.0.0",
		Compiled:  time.Now(),
		Authors:   []*cli.Author{{Name: "Doug Evenhouse", Email: "doug@evenhouseconsulting.com"}},
		Copyright: "(c) 2024 Evenhouse Consulting, Inc.",
		Commands: []*cli.Command{
			newUploadCommand(log),
			newDownloadCommand(log),
			newPruneCommand(log),
		},
	}
	return app
}

func newUploadCommand(log *logger.Logger) *cli.Command {
	return &cli.Command{
		Name:      "upload",
		Aliases:   []string{"u"},
		Usage:     "Upload backup files to cloud storage",
		ArgsUsage: "[root directory]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "non-interactive",
				Usage: "Run in non-interactive mode (no progress bars)",
			},
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "Storage provider (s3 or b2)",
				Value:   "s3",
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region (for S3 only)",
				Value:   "us-east-1",
			},
			&cli.StringFlag{
				Name:     "bucket",
				Aliases:  []string{"b"},
				Usage:    "Storage bucket name",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "keyprefix",
				Aliases: []string{"k"},
				Usage:   "Prefix for storage keys",
			},
			&cli.StringFlag{
				Name:    "backupext",
				Aliases: []string{"x"},
				Usage:   "File extension for backup files",
				Value:   ".bak",
			},
			&cli.BoolFlag{
				Name:    "compress",
				Aliases: []string{"c"},
				Usage:   "Compress files before uploading",
			},
		},
		Action: func(c *cli.Context) error {
			uploader, err := getUploader(c)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
			return storage.Upload(c, uploader, log)
		},
	}
}

func newDownloadCommand(log *logger.Logger) *cli.Command {
	return &cli.Command{
		Name:      "download",
		Aliases:   []string{"d"},
		Usage:     "Download backup files from cloud storage",
		ArgsUsage: "[key]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "Storage provider (s3 or b2)",
				Value:   "s3",
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region (for S3 only)",
				Value:   "us-east-1",
			},
			&cli.StringFlag{
				Name:     "bucket",
				Aliases:  []string{"b"},
				Usage:    "Storage bucket name",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file name",
			},
		},
		Action: func(c *cli.Context) error {
			uploader, err := getUploader(c)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
			return storage.Download(c, uploader, log)
		},
	}
}

func newPruneCommand(log *logger.Logger) *cli.Command {
	return &cli.Command{
		Name:    "prune",
		Aliases: []string{"p"},
		Usage:   "Remove old backup files from cloud storage",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "Storage provider (s3 or b2)",
				Value:   "s3",
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region (for S3 only)",
				Value:   "us-east-1",
			},
			&cli.StringFlag{
				Name:     "bucket",
				Aliases:  []string{"b"},
				Usage:    "Storage bucket name",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "keyprefix",
				Aliases: []string{"k"},
				Usage:   "Prefix for storage keys",
			},
			&cli.DurationFlag{
				Name:     "age",
				Aliases:  []string{"a"},
				Usage:    "Age of files to prune (e.g., 720h for 30 days)",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			uploader, err := getUploader(c)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
			return storage.Prune(c, uploader, log)
		},
	}
}

func getUploader(c *cli.Context) (storage.Uploader, error) {
	provider := c.String("provider")
	bucket := c.String("bucket")

	switch provider {
	case "s3":
		region := c.String("region")
		return storage.NewS3Uploader(region, bucket)
	case "b2":
		return storage.NewB2Uploader(bucket)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", provider)
	}
}