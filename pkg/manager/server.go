package manager

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

type jobServer struct {
	K8sClient client.Client
	Addr      string
}

func (s jobServer) Start(ctx context.Context) error {
	m := http.NewServeMux()
	m.Handle("/jobs/start", handlers.StartJob{K8sClient: s.K8sClient})
	m.Handle("/jobs/status", handlers.GetJobStatus{K8sClient: s.K8sClient})

	logger := ctrl.LoggerFrom(ctx).WithName("job-server")
	ctrl.LoggerInto(ctx, logger)
	sv := http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
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
