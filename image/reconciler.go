package image

import (
	"context"

	eiriniv1 "code.cloudfoundry.org/eirini/pkg/apis/eirini/v1"
	"code.cloudfoundry.org/lager"
	kpackv1alphav1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconciler(logger lager.Logger) *Reconciler {
	return &Reconciler{
		logger: logger,
	}
}

type Reconciler struct {
	logger        lager.Logger
	runtimeClient client.Client
}

func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := r.logger.Session("reconcile-image",
		lager.Data{
			"name":      request.NamespacedName.Name,
			"namespace": request.NamespacedName.Namespace,
		})

	image := kpackv1alphav1.Image{}
	if err := r.runtimeClient.Get(context.Background(), request.NamespacedName, &image); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("image-not-found", err)
			return reconcile.Result{}, nil
		}

		logger.Error("failed-to-get-image", err)
		return reconcile.Result{}, errors.Wrap(err, "failed to get image")
	}

	if !imageIsReady(image) {
		logger.Debug("image not yet ready")
		return reconcile.Result{}, nil
	}

	err := r.desireLRP(image)
	if err != nil {
		logger.Error("failed-to-desire-lrp", err)
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) desireLRP(image kpackv1alphav1.Image) error {
	lrp := &eiriniv1.LRP{
		ObjectMeta: metav1.ObjectMeta{
			Name: "carl",
		},
		Spec: eiriniv1.LRPSpec{
			Image:   image.Spec.Source.Registry.Image,
			DiskMB:  512,
			AppGUID: "aaaap-guid",
			GUID:    "much-guid",
			Version: "v1",
		},
	}
	r.runtimeClient.Create(context.Background(), lrp)
	return nil
}

func imageIsReady(image kpackv1alphav1.Image) bool {
	return image.Spec.Source.Registry != nil || image.Spec.Source.Registry.Image == ""
}
