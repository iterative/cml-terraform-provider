package machine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"time"

	units "github.com/docker/go-units"

	_ "github.com/rclone/rclone/backend/azureblob"
	_ "github.com/rclone/rclone/backend/googlecloudstorage"
	_ "github.com/rclone/rclone/backend/local"
	_ "github.com/rclone/rclone/backend/s3"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/accounting"
	"github.com/rclone/rclone/fs/filter"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rclone/rclone/fs/sync"

	"github.com/sirupsen/logrus"

	"terraform-provider-iterative/task/common"
)

type StatusReport struct {
	Result string
	Status string
	Code   string
}

func init() {
	operations.SyncPrintf = func(format string, a ...interface{}) {
		logrus.Debugf(format, a...)
	}
	fs.LogPrint = func(level fs.LogLevel, text string) {
		logrus.Debug(text)
	}
}

func Reports(ctx context.Context, remote, prefix string) ([]string, error) {
	remoteFileSystem, err := fs.NewFs(ctx, remote)
	if err != nil {
		return nil, err
	}

	entries, err := remoteFileSystem.List(ctx, "/reports")
	if err != nil {
		return nil, err
	}

	var logs []string
	for _, entry := range entries {
		path := entry.Remote()
		if base := filepath.Base(path); !strings.HasPrefix(base, prefix+"-") {
			continue
		}

		object, err := remoteFileSystem.NewObject(ctx, path)
		if err != nil {
			return nil, err
		}
		reader, err := object.Open(ctx)
		if err != nil {
			return nil, err
		}
		buffer := new(bytes.Buffer)
		if _, err := io.Copy(buffer, reader); err != nil {
			return nil, err
		}
		logs = append(logs, buffer.String())
		reader.Close()
	}

	return logs, nil
}

func Logs(ctx context.Context, remote string) ([]string, error) {
	return Reports(ctx, remote, "task")
}

func Status(ctx context.Context, remote string, initialStatus common.Status) (common.Status, error) {
	reports, err := Reports(ctx, remote, "status")
	if err != nil {
		return initialStatus, err
	}

	for _, report := range reports {
		var statusReport StatusReport
		if err := json.Unmarshal([]byte(report), &statusReport); err != nil {
			return initialStatus, err
		}
		if statusReport.Code != "" {
			if statusReport.Code == "0" {
				initialStatus[common.StatusCodeSucceeded] += 1
			} else {
				initialStatus[common.StatusCodeFailed] += 1
			}
		} else if statusReport.Result == "timeout" {
			initialStatus[common.StatusCodeFailed] += 1
		}
	}
	return initialStatus, nil
}

func Transfer(ctx context.Context, source, destination string, include string) error {
	if include = filepath.Clean(include); filepath.IsAbs(include) || strings.HasPrefix(include, "../") {
		return errors.New("storage.output must be inside storage.workdir")
	}

	rules := []string{
		"+ " + filepath.Clean("/"+include),
		"+ " + filepath.Clean("/"+include+"/**"),
		"- **",
	}

	ctx, fi := filter.AddConfig(ctx)
	for _, rule := range rules {
		if err := fi.AddRule(rule); err != nil {
			return err
		}
	}

	sourceFileSystem, err := fs.NewFs(ctx, source)
	if err != nil {
		return err
	}

	destinationFileSystem, err := fs.NewFs(ctx, destination)
	if err != nil {
		return err
	}

	if count, size, err := operations.Count(ctx, sourceFileSystem); err == nil {
		logrus.Infof("Transferring %s (%d files)...", units.HumanSize(float64(size)), count)
	} else {
		return err
	}

	defer progress(10 * time.Second)()

	return sync.CopyDir(ctx, destinationFileSystem, sourceFileSystem, true)
}

func Delete(ctx context.Context, destination string) error {
	destinationFileSystem, err := fs.NewFs(ctx, destination)
	if err != nil {
		return err
	}

	actions := []func(context.Context) error{
		func(ctx context.Context) error {
			return operations.Delete(ctx, destinationFileSystem)
		},
		func(ctx context.Context) error {
			return operations.Rmdirs(ctx, destinationFileSystem, "", true)
		},
	}

	for _, action := range actions {
		if err := action(ctx); err != nil {
			if !errors.Is(err, fs.ErrorDirNotFound) && !strings.Contains(err.Error(), "no such host") {
				return common.NotFoundError
			}
			return err
		}
	}

	return nil
}

func progress(interval time.Duration) func() {
	accounting.GlobalStats().ResetCounters()
	ci := fs.GetConfig(context.Background())
	ci.StatsOneLine = true

	ticker := time.NewTicker(interval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				logrus.Info(accounting.GlobalStats().String())
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		done <- true
	}
}
