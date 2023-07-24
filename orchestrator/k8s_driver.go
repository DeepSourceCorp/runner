package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const DefaultK8sTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

type K8sDriver struct {
	clientset *kubernetes.Clientset
}

func NewK8sDriver(tokenPath string) (Driver, error) {
	if tokenPath == "" {
		tokenPath = DefaultK8sTokenPath
	}
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}
	// Create a Kubernetes REST client using the service account token
	config := &rest.Config{
		Host:        fmt.Sprintf("https://%s", os.Getenv("KUBERNETES_SERVICE_HOST")),
		BearerToken: string(token),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sDriver{clientset: clientset}, nil
}

// TriggerJob creates the kubernetes job supplied as a parameter.
func (d *K8sDriver) TriggerJob(ctx context.Context, job JobCreator) error {
	k8sJob := &MarvinK8sJob{job}
	j, err := k8sJob.Job()
	if err != nil {
		return err
	}
	_, err = d.clientset.BatchV1().Jobs(k8sJob.Namespace()).Create(ctx, j, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteJob deletes the kubernetes job supplied as a parameter.
func (d *K8sDriver) DeleteJob(ctx context.Context, job JobDeleter) error {
	foregroundDeletion := metav1.DeletePropagationForeground
	log.Println("Deleting job", job.Name())
	err := d.clientset.BatchV1().Jobs(job.Namespace()).Delete(ctx, job.Name(), metav1.DeleteOptions{
		PropagationPolicy: &foregroundDeletion,
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *K8sDriver) CleanExpiredJobs(ctx context.Context, namespace string, interval *time.Duration) error {
	// set the propagation policy to foreground
	deletePropagationPolicy := metav1.DeletePropagationForeground
	// get a list of all jobs in the namespace
	jobs, err := d.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "manager=runner",
	})
	if err != nil {
		return err
	}
	// delete all completed and failed jobs
	for _, job := range jobs.Items {
		if (job.Status.Succeeded > 0 || job.Status.Failed > 0) && job.Status.StartTime.Before(&metav1.Time{Time: time.Now().Add(*interval)}) {
			err := d.clientset.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{
				PropagationPolicy: &deletePropagationPolicy,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
