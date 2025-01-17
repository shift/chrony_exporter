// Copyright 2022 Ben Kochie
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"net"

	"github.com/facebook/time/ntp/chrony"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	trackingSubsystem = "tracking"
)

var (
	// The remote IP 127.127.1.1 means it is a "local" reference clock.
	trackingLocalIP = net.IPv4(127, 127, 1, 1)

	trackingInfo = typedDesc{
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, trackingSubsystem, "info"),
			"Chrony tracking info",
			[]string{"tracking_refid", "tracking_address"},
			nil,
		),
		prometheus.GaugeValue,
	}

	trackingLastOffset = typedDesc{
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, trackingSubsystem, "last_offset_seconds"),
			"Chrony tracking last offset in seconds",
			nil,
			nil,
		),
		prometheus.GaugeValue,
	}

	trackingRefTime = typedDesc{
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, trackingSubsystem, "reference_timestamp_seconds"),
			"Chrony tracking Refreence timestamp",
			nil,
			nil,
		),
		prometheus.GaugeValue,
	}

	trackingRemoteTracking = typedDesc{
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, trackingSubsystem, "remote_reference"),
			"Chrony tracking is connected to a remote source",
			nil,
			nil,
		),
		prometheus.GaugeValue,
	}

	trackingRMSOffset = typedDesc{
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, trackingSubsystem, "rms_offset_seconds"),
			"Chrony tracking long-term average of the offset",
			nil,
			nil,
		),
		prometheus.GaugeValue,
	}

	trackingStratum = typedDesc{
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, trackingSubsystem, "stratum"),
			"Chrony tracking client stratum",
			nil,
			nil,
		),
		prometheus.GaugeValue,
	}
)

func (e Exporter) getTrackingMetrics(ch chan<- prometheus.Metric, client chrony.Client) {
	packet, err := client.Communicate(chrony.NewTrackingPacket())
	if err != nil {
		level.Error(e.logger).Log("msg", "Couldn't get tracking", "err", err)
		return
	}
	level.Debug(e.logger).Log("msg", "Got 'tracking' response", "tracking_packet", packet.GetStatus())

	tracking, ok := packet.(*chrony.ReplyTracking)
	if !ok {
		level.Error(e.logger).Log("msg", "Got wrong 'tracking' response", "packet", packet)
		return
	}

	ch <- trackingInfo.mustNewConstMetric(1.0, chrony.RefidToString(tracking.RefID), tracking.IPAddr.String())

	ch <- trackingLastOffset.mustNewConstMetric(tracking.LastOffset)
	level.Debug(e.logger).Log("msg", "Tracking Last Offset", "offset", tracking.LastOffset)

	ch <- trackingRefTime.mustNewConstMetric(float64(tracking.RefTime.UnixNano()) / 1e9)
	level.Debug(e.logger).Log("msg", "Tracking Ref Time", "ref_time", tracking.RefTime)

	remoteTracking := 1.0
	if tracking.IPAddr.Equal(trackingLocalIP) {
		remoteTracking = 0.0
	}
	ch <- trackingRemoteTracking.mustNewConstMetric(remoteTracking)
	level.Debug(e.logger).Log("msg", "Tracking is remote", "bool_value", remoteTracking)

	ch <- trackingRMSOffset.mustNewConstMetric(tracking.RMSOffset)
	level.Debug(e.logger).Log("msg", "Tracking RMS Offset", "rms_offset", tracking.RMSOffset)

	ch <- trackingStratum.mustNewConstMetric(float64(tracking.Stratum))
	level.Debug(e.logger).Log("msg", "Tracking Stratum", "stratum", tracking.Stratum)
}
