package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

type StartJob struct {
	K8sClient client.Client
}

func (s StartJob) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := ctrl.LoggerFrom(ctx)
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %s", err), 400)
		return
	}
	namespace := r.Form.Get("namespace")
	if namespace == "" {
		http.Error(w, "namespace is required", 400)
		return
	}
	name := r.Form.Get("name")
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

type startJobOutput struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

func (s StartJob) handle(ctx context.Context, namespace, name string) (*startJobOutput, error) {
	logger := ctrl.LoggerFrom(ctx)

	var cronJob batchv1.CronJob
	if err := s.K8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, &cronJob); err != nil {
		return nil, fmt.Errorf("could not get CronJob: %w", err)
	}
	logger.Info("found CronJob", "cronJob", cronJob.ObjectMeta)

	jobSpec := cronJob.Spec.JobTemplate.Spec.DeepCopy()
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    cronJob.Namespace,
			GenerateName: fmt.Sprintf("%s-", cronJob.Name),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: cronJob.APIVersion,
				Kind:       cronJob.Kind,
				Name:       cronJob.Name,
				UID:        cronJob.UID,
				Controller: pointer.Bool(true),
			}},
		},
		Spec: *jobSpec,
	}
	if err := s.K8sClient.Create(ctx, &job); err != nil {
		return nil, fmt.Errorf("could not create Job: %w", err)
	}
	logger.Info("created Job", "job", job.ObjectMeta)

	return &startJobOutput{
		Namespace: job.Namespace,
		Name:      job.Name,
	}, nil
}
