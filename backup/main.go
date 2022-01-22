package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
)

func main() {
	// read config
	cfg := parseConfig()

	// build logger
	l, err := zapdriver.NewProduction()
	if err != nil {
		panic(err)
	}
	l = l.With(
		zap.String("bucket", cfg.Bucket),
		zap.String("prefix", cfg.Prefix),
	)

	// create program context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// get cloud storage client
	c, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	bkt := c.Bucket(cfg.Bucket)

	// create pg_dumpall command
	cmd := exec.CommandContext(ctx, "pg_dumpall", cfg.Args...)

	// start main loop
	for {
		// calculate next backup time and create object handle
		next := nextBackup(cfg.Schedule.Hour, cfg.Schedule.Minute)
		obj := bkt.Object(cfg.Prefix + next.Format(time.RFC3339) + ".sql.gz")
		l.Info("next backup", zap.Time("time", next))

		// wait until next backup time and run backup
		select {
		case <-time.After(time.Until(next)):
			if err := runBackup(ctx, l, cmd, obj); err != nil {
				l.Error("error running backup", zap.Error(err))
				return
			}
		case <-sigs:
			l.Warn("shutting down")
			return // exit if term signal received
		}
	}
}

func runBackup(ctx context.Context, l *zap.Logger, cmd *exec.Cmd, obj *storage.ObjectHandle) error {
	// create writer to cloud storage object
	gcsw := obj.NewWriter(ctx)
	defer func() {
		if err := gcsw.Close(); err != nil {
			l.Error("close object writer", zap.Error(err))
		}
	}()

	// compress data written to object
	gzw, err := gzip.NewWriterLevel(gcsw, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("create gzip writer: %w", err)
	}
	defer func() {
		if err := gzw.Close(); err != nil {
			l.Error("close gzip writer", zap.Error(err))
		}
	}()

	l.Info("running backup")

	// pg_dumpall will write backup to stdout
	cmd.Stdout = gzw
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	l.Info("backup complete")
	return nil
}

func nextBackup(hour, minute int) time.Time {
	now := time.Now()

	next := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		hour, minute, 0, 0,
		now.Location(),
	)

	// if the next backup is later today, return
	if next.After(now) {
		return next
	}

	// next backup is tomorrow
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day()+1,
		hour, minute, 0, 0,
		now.Location(),
	)
}
