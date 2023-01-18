package metrics

import (
	"context"
	"time"

	"github.com/appliedres/cloudy/vm"
)

type Metric[T any] struct {
	Timestamp time.Time
	Value     T
}

func NewMetric[T any](ctx context.Context, value T) *Metric[T] {
	return &Metric[T]{
		Timestamp: time.Now(),
		Value:     value,
	}
}

type MetricsRecorder interface {
	RecordVMStatus(ctx context.Context, metric *Metric[*vm.VirtualMachineStatus]) error
}

type NoOpMetricRecorder struct{}

func NewNoOpMetricRecorder() *NoOpMetricRecorder {
	return &NoOpMetricRecorder{}
}

func (rec *NoOpMetricRecorder) RecordVMStatus(ctx context.Context, metric *Metric[*vm.VirtualMachineStatus]) error {
	return nil
}
