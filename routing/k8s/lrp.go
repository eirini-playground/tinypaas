package k8s

import (
	"context"

	eiriniv1 "code.cloudfoundry.org/eirini/pkg/apis/eirini/v1"
	"code.cloudfoundry.org/lager"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewRoutingReconciler(logger lager.Logger, lrpsClient runtimeclient.Client, clientset kubernetes.Interface, scheme *runtime.Scheme) *RoutingReconciler {
	return &RoutingReconciler{
		logger:     logger,
		lrpsClient: lrpsClient,
		clientset:  clientset,
		scheme:     scheme,
	}
}

type RoutingReconciler struct {
	logger     lager.Logger
	lrpsClient runtimeclient.Client
	clientset  kubernetes.Interface
	scheme     *runtime.Scheme
}

func (r *RoutingReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := r.logger.Session("reconcile-lrp",
		lager.Data{
			"name":      request.NamespacedName.Name,
			"namespace": request.NamespacedName.Namespace,
		})

	lrp := &eiriniv1.LRP{}
	if err := r.lrpsClient.Get(context.Background(), request.NamespacedName, lrp); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("lrp-not-found", err)

			return reconcile.Result{}, nil
		}

		logger.Error("failed-to-get-lrp", err)

		return reconcile.Result{}, errors.Wrap(err, "failed to get lrp")
	}

	err := r.do(lrp)
	if err != nil {
		logger.Error("failed-to-reconcile", err)
	}

	return reconcile.Result{}, err
}

func (r *RoutingReconciler) do(lrp *eiriniv1.LRP) error {
	r.logger.Info("reconciling lrp for app " + lrp.Spec.AppName)
	return nil
}
