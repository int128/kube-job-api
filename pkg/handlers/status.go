package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

type GetJobStatus struct {
	K8sClient client.Client
}

func (s GetJobStatus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := ctrl.LoggerFrom(ctx)
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	namespace := query.Get("namespace")
	if namespace == "" {
		http.Error(w, "namespace is required", 400)
		return
	}
	name := query.Get("name")
	if namespace == "" {
		http.Error(w, "name is required", 400)
		return
	}

	logger = logger.WithValues("namespace", namespace, "name", name)
	output, err := s.handle(ctx, namespace, name)
	if err != nil {
		logger.Error(err, "handler error")
		http.Error(w, fmt.Sprintf("error: %s", err), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(output); err != nil {
		logger.Error(err, "could not write json")
	}
}

type getJobStatusOutput struct {
	batchv1.JobStatus
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func (s GetJobStatus) handle(ctx context.Context, namespace, name string) (*getJobStatusOutput, error) {
	logger := ctrl.LoggerFrom(ctx)

	var job batchv1.Job
	if err := s.K8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &job); err != nil {
		return nil, fmt.Errorf("could not get CronJob: %w", err)
	}
	logger.Info("found Job", "status", job.Status)

	return &getJobStatusOutput{
		JobStatus: job.Status,
		Namespace: job.Namespace,
		Name:      job.Name,
	}, nil
}
