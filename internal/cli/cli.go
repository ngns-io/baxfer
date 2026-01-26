package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/ngns-io/baxfer/pkg/logger"
	"github.com/ngns-io/baxfer/pkg/storage"
	"github.com/urfave/cli/v2"
)

var version = "dev" // Will be overwritten by build flag

func NewApp() *cli.App {
	app := &cli.App{
		Name:      "baxfer",
		Usage:     "CLI to help manage storage for database backups",
		Version:   version,
		Compiled:  time.Now(),
		Authors:   []*cli.Author{{Name: "Doug Evenhouse", Email: "doug@evenhouseconsulting.com"}},
		Copyright: "(c) 2024 Evenhouse Consulting, Inc.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "logfile",
				Aliases: []string{"l"},
				Usage:   "Specify the log file location",
				Value:   "baxfer.log",
			},
			&cli.IntFlag{
				Name:  "log-max-size",
				Usage: "Maximum size of log file before rotation (in megabytes)",
				Value: 10,
			},
			&cli.IntFlag{
				Name:  "log-max-age",
				Usage: "Maximum number of days to retain old log files",
				Value: 30,
			},
			&cli.IntFlag{
				Name:  "log-max-backups",
				Usage: "Maximum number of old log files to retain",
				Value: 5,
			},
			&cli.BoolFlag{
				Name:  "log-compress",
				Usage: "Compress rotated log files",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "log-clear",
				Usage: "Clear log file on start",
				Value: false,
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Enable quiet mode (log only errors)",
			},
		},
		Commands: []*cli.Command{
			newUploadCommand(),
			newDownloadCommand(),
			newPruneCommand(),
		},
		Before: func(c *cli.Context) error {
			// Initialize logger
			logConfig := logger.LogConfig{
				Filename:     c.String("logfile"),
				MaxSize:      c.Int("log-max-size"),
				MaxAge:       c.Int("log-max-age"),
				MaxBackups:   c.Int("log-max-backups"),
				Compress:     c.Bool("log-compress"),
				ClearOnStart: c.Bool("log-clear"),
			}
			log, err := logger.New(logConfig, c.Bool("quiet"))
			if err != nil {
				return err
			}
			c.App.Metadata["logger"] = log
			return nil
		},
		After: func(c *cli.Context) error {
			// Close logger
			log, err := getLogger(c)
			if err != nil {
				return err
			}
			return log.Close()
		},
	}
	return app
}

func newUploadCommand() *cli.Command {
	cmd := &cli.Command{
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
				Usage:   "Storage provider (s3, b2, b2s3, r2, or sftp)",
				Value:   "s3",
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region (for S3 and b2s3 only)",
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
			log, err := getLogger(c)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			uploader, err := getUploader(c)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return storage.Upload(c, uploader, log)
		},
	}
	cmd.Flags = append(cmd.Flags, sftpFlags()...)
	return cmd
}

func newDownloadCommand() *cli.Command {
	cmd := &cli.Command{
		Name:      "download",
		Aliases:   []string{"d"},
		Usage:     "Download backup files from cloud storage",
		ArgsUsage: "[key]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "Storage provider (s3, b2, b2s3, r2, or sftp)",
				Value:   "s3",
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region (for S3 and b2s3 only)",
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
			log, err := getLogger(c)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			uploader, err := getUploader(c)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			err = storage.Download(c, uploader, log)
			if err != nil {
				// If it's our user error, just show the message
				var userErr *storage.UserError
				if errors.As(err, &userErr) {
					return cli.Exit(userErr.Message, 1)
				}
				// For unexpected errors, show a generic message
				return cli.Exit("An unexpected error occurred while downloading the file", 1)
			}
			return nil
		},
	}
	cmd.Flags = append(cmd.Flags, sftpFlags()...)
	return cmd
}

func newPruneCommand() *cli.Command {
	cmd := &cli.Command{
		Name:    "prune",
		Aliases: []string{"p"},
		Usage:   "Remove old backup files from cloud storage",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "Storage provider (s3, b2, b2s3, r2, or sftp)",
				Value:   "s3",
			},
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Usage:   "AWS region (for S3 and b2s3 only)",
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
			log, err := getLogger(c)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			uploader, err := getUploader(c)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return storage.Prune(c, uploader, log)
		},
	}
	cmd.Flags = append(cmd.Flags, sftpFlags()...)
	return cmd
}

func sftpFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "sftp-host",
			Usage:   "SFTP server hostname",
			EnvVars: []string{"SFTP_HOST"},
		},
		&cli.IntFlag{
			Name:    "sftp-port",
			Usage:   "SFTP server port",
			Value:   22,
			EnvVars: []string{"SFTP_PORT"},
		},
		&cli.StringFlag{
			Name:    "sftp-user",
			Usage:   "SFTP username",
			EnvVars: []string{"SFTP_USER"},
		},
		&cli.StringFlag{
			Name:    "sftp-path",
			Usage:   "Base path on SFTP server",
			EnvVars: []string{"SFTP_PATH"},
		},
	}
}

func getLogger(c *cli.Context) (logger.Logger, error) {
	log, ok := c.App.Metadata["logger"].(logger.Logger)
	if !ok {
		return nil, fmt.Errorf("logger not initialized")
	}
	return log, nil
}

func getUploader(c *cli.Context) (storage.Uploader, error) {
	provider := c.String("provider")
	bucket := c.String("bucket")
	log, err := getLogger(c)
	if err != nil {
		return nil, err
	}

	switch provider {
	case "s3":
		region := c.String("region")
		return storage.NewS3Uploader(region, bucket, log)
	case "b2":
		return storage.NewB2Uploader(bucket, log)
	case "b2s3":
		region := c.String("region")
		return storage.NewB2S3Uploader(region, bucket, log)
	case "r2":
		return storage.NewR2Uploader(bucket, log)
	case "sftp":
		host := c.String("sftp-host")
		port := c.Int("sftp-port")
		user := c.String("sftp-user")
		path := c.String("sftp-path")
		if host == "" || user == "" || path == "" {
			return nil, fmt.Errorf("SFTP provider requires host, user, and path")
		}
		return storage.NewSFTPUploader(host, port, user, path, log)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", provider)
	}
}
