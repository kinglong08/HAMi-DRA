/*
Copyright 2025 The HAMi Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"testing"

	"github.com/Project-HAMi/HAMi-DRA/pkg/cache"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes/fake"
)

func newTestCollector(withDevice bool) *Collector {
	c := cache.NewCacheWithClient(fake.NewSimpleClientset())
	if withDevice {
		c.NodeDevices.Nodes["node1"] = &cache.NodeDeviceInfo{
			Devices: []*cache.NodeDevice{
				{Name: "gpu0", UUID: "uuid-1", Brand: "NVIDIA", ProductName: "V100"},
			},
		}
	}
	c.NodeDevices.Claims.Claims["default/claim1"] = &cache.DeviceAllocation{
		NodeName: "node1",
		UsedBy:   []string{"pod1"},
		AllocationResults: []*cache.AllocationResult{
			{Namespace: "default", DeviceName: "gpu0", Cores: 50, Memory: 8192},
		},
	}
	return NewCollector(c)
}

func collectAll(col *Collector) int {
	ch := make(chan prometheus.Metric, 20)
	col.Collect(ch)
	close(ch)
	n := 0
	for range ch {
		n++
	}
	return n
}

func TestCollect_WithDevice(t *testing.T) {
	// 4 node metrics + 2 pod metrics (core + memory per pod)
	if got := collectAll(newTestCollector(true)); got != 6 {
		t.Errorf("expected 6 metrics, got %d", got)
	}
}

func TestCollect_MissingDevice(t *testing.T) {
	// device not in cache, pod metrics are skipped
	if got := collectAll(newTestCollector(false)); got != 0 {
		t.Errorf("expected 0 metrics, got %d", got)
	}
}
