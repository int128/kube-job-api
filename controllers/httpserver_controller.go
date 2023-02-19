/*
Copyright 2022.

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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/int128/kube-job-server/pkg/handlers"
	"github.com/int128/kube-job-server/static"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HTTPServerController manages a HTTPServer
type HTTPServerController struct {
	client.Client
	Scheme *runtime.Scheme
	Addr   string
}

// Start starts the HTTPServer
func (r *HTTPServerController) Start(ctx context.Context) error {
	m := http.NewServeMux()
	m.Handle("/api/jobs/start", handlers.StartJob{K8sClient: r.Client})
	m.Handle("/api/jobs/status", handlers.GetJobStatus{K8sClient: r.Client})
	m.Handle("/", http.FileServer(http.FS(static.FS())))

	logger := ctrl.LoggerFrom(ctx).WithName("job-server")
	ctrl.LoggerInto(ctx, logger)
	sv := http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		Addr:        r.Addr,
		Handler:     m,
	}
	go func() {
		<-ctx.Done()
		logger.Info("Stopping server")
		if err := sv.Close(); err != nil {
			logger.Error(err, "could not close server")
		}
	}()
	logger.Info("Starting server", "addr", sv.Addr)
	if err := sv.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("could not start http server: %w", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HTTPServerController) SetupWithManager(mgr ctrl.Manager) error {
	return mgr.Add(r)
}
