package orchestrator

import (
	"context"
	"os"
	"time"

	"k8s.io/cli-runtime/pkg/printers"
)

const (
	DriverPrinter = "printer"
)

type K8sPrinterDriver struct{}

func NewK8sPrinterDriver() Driver {
	return &K8sPrinterDriver{}
}

func (*K8sPrinterDriver) TriggerJob(_ context.Context, job JobCreator) error {
	k8sJob := &MarvinK8sJob{job}
	j, err := k8sJob.Job()
	if err != nil {
		return err
	}
	printer := printers.YAMLPrinter{}
	return printer.PrintObj(j, os.Stdout)
}

func (*K8sPrinterDriver) DeleteJob(_ context.Context, _ JobDeleter) error {
	return nil
}

func (*K8sPrinterDriver) CleanExpiredJobs(_ context.Context, _ string, _ *time.Duration) error {
	return nil
}
