package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/int128/kube-job-server/pkg/handlers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Server struct {
	K8sClient client.Client
	Addr      string
}

func (s Server) Start(ctx context.Context) error {
	logger := ctrl.LoggerFrom(ctx).WithName("http-server")
	ctrl.LoggerInto(ctx, logger)

	m := http.NewServeMux()
	m.Handle("/jobs/start", handlers.StartJob{K8sClient: s.K8sClient})
	m.Handle("/jobs/status", handlers.GetJobStatus{K8sClient: s.K8sClient})

	sv := http.Server{
		BaseContext: func(listener net.Listener) context.Context { return ctx },
		Addr:        s.Addr,
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
