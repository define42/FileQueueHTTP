package prometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FileAddedToChannel = promauto.NewCounter(prometheus.CounterOpts{
		Name: "file_added_to_channel",
		Help: "The total number of files added to channel",
	})

	FileInQueue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "file_in_queue",
		Help: "The numbers of files in queue",
	})
	FilesDoNotExist = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "files_do_not_exist",
		Help: "The numbers of files do not exist",
	})
)

var ctx = context.Background()

func SystemStat() {
	/*
		for {
			usageStat, err := disk.UsageWithContext(ctx, config.Settings.Get(config.DATA_FOLDER))
			if err == nil {
				Disk_used_percent.Set(float64(usageStat.UsedPercent))
				Disk_free.Set(float64(usageStat.Free))
			}

			time.Sleep(1000 * time.Millisecond)
		}
	*/
}
