package image

import (
	"context"

	eiriniv1 "code.cloudfoundry.org/eirini/pkg/apis/eirini/v1"
	"code.cloudfoundry.org/lager"
	kpackv1alphav1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackscheme "github.com/pivotal/kpack/pkg/client/clientset/versioned/scheme"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconciler(logger lager.Logger, runtimeClient client.Client) *Reconciler {
	return &Reconciler{
		logger:        logger,
		runtimeClient: runtimeClient,
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
	logger.Debug("reconcile started")
	defer logger.Debug("reconcile finished")

	logger.Debug("getting image")
	image := kpackv1alphav1.Image{}
	if err := r.runtimeClient.Get(context.Background(), request.NamespacedName, &image); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("image-not-found", lager.Data{"error": err})
			return reconcile.Result{}, nil
		}

		logger.Error("failed-to-get-image", err)
		return reconcile.Result{}, errors.Wrap(err, "failed to get image")
	}

	logger.Debug("checking image readiness")
	if !imageIsReady(image) {
		logger.Debug("image not yet ready")
		return reconcile.Result{}, nil
	}

	namespacedName := types.NamespacedName{
		Name:      image.Name,
		Namespace: image.Namespace,
	}

	lrp := eiriniv1.LRP{}
	if err := r.runtimeClient.Get(context.Background(), namespacedName, &lrp); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("lrp-not-found", lager.Data{"error": err})

			logger.Debug("desiring lrp")
			err := r.desireLRP(image)
			if err != nil {
				logger.Error("failed-to-desire-lrp", err)
				return reconcile.Result{}, errors.Wrap(err, "failed to desire lrp")
			}
			return reconcile.Result{}, nil
		}

		logger.Error("failed-to-get-lrp", err)
		return reconcile.Result{}, errors.Wrap(err, "failed to get image")
	}

	logger.Debug("updating lrp")
	if image.Status.LatestImage != lrp.Spec.Image {
		lrp.Spec.Image = image.Status.LatestImage
		if err := r.runtimeClient.Update(context.Background(), &lrp); err != nil {
			logger.Error("failed-to-update-lrp", err)
			return reconcile.Result{}, errors.Wrap(err, "failed to update lrp")
		}

	}
	return reconcile.Result{}, nil
}

func (r *Reconciler) desireLRP(image kpackv1alphav1.Image) error {
	lrp := &eiriniv1.LRP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      image.Name,
			Namespace: "eirini-workloads",
		},
		Spec: eiriniv1.LRPSpec{
			Image:     image.Status.LatestImage,
			Instances: 1,
			AppGUID:   image.Name,
			GUID:      image.Name,
			Version:   "v1",
			DiskMB:    512,
			AppRoutes: []eiriniv1.Route{},
		},
	}

	if err := ctrl.SetControllerReference(&image, lrp, kpackscheme.Scheme); err != nil {
		return err
	}

	return r.runtimeClient.Create(context.Background(), lrp)
}

func imageIsReady(image kpackv1alphav1.Image) bool {
	for _, condition := range image.Status.Conditions {
		if condition.Type == "Ready" && condition.Status == "True" {
			return true
		}
	}
	return false
}
