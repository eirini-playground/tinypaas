package k8s

import (
	"context"
	"fmt"

	eirinik8s "code.cloudfoundry.org/eirini/k8s"
	eiriniv1 "code.cloudfoundry.org/eirini/pkg/apis/eirini/v1"
	"code.cloudfoundry.org/lager"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
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
			logger.Info("lrp-not-found", lager.Data{"namespace": request.Namespace, "name": request.Name})

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

	_, err := r.getOrCreateService(lrp)
	if err != nil {
		return err
	}

	ingress, err := r.getOrCreateIngress(lrp)
	if err != nil {
		return err
	}
	ingress.Spec.Rules = r.createIngressRules(lrp)
	_, err = r.clientset.ExtensionsV1beta1().Ingresses(lrp.Namespace).Update(context.Background(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (r *RoutingReconciler) generateServiceSpec(lrp *eiriniv1.LRP) *corev1.Service {
	servicePorts := []corev1.ServicePort{}
	for _, port := range lrp.Spec.Ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Port:       port,
			TargetPort: intstr.FromInt(int(port)),
		})
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lrp.Spec.GUID,
			Namespace: lrp.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: servicePorts,
			Selector: map[string]string{
				eirinik8s.LabelGUID: lrp.Spec.GUID,
			},
		},
	}

	return service
}

func (r *RoutingReconciler) getOrCreateService(lrp *eiriniv1.LRP) (*corev1.Service, error) {
	var service *corev1.Service
	var err error
	service, err = r.clientset.CoreV1().Services(lrp.Namespace).Get(context.Background(), lrp.Spec.GUID, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			service, err = r.createService(lrp)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return service, nil
}

func (r *RoutingReconciler) getOrCreateIngress(lrp *eiriniv1.LRP) (*v1beta1.Ingress, error) {
	var ingress *v1beta1.Ingress
	var err error
	ingress, err = r.clientset.ExtensionsV1beta1().Ingresses(lrp.Namespace).Get(context.Background(), lrp.Spec.GUID, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			ingress, err = r.createIngress(lrp)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return ingress, nil
}

func (r *RoutingReconciler) createIngress(lrp *eiriniv1.LRP) (*v1beta1.Ingress, error) {
	ingress := r.generateIngressSpec(lrp)
	if err := ctrl.SetControllerReference(lrp, ingress, r.scheme); err != nil {
		return nil, err
	}

	return r.clientset.ExtensionsV1beta1().Ingresses(lrp.Namespace).Create(
		context.Background(),
		ingress,
		metav1.CreateOptions{},
	)
}

func (r *RoutingReconciler) generateIngressSpec(lrp *eiriniv1.LRP) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lrp.Spec.GUID,
			Namespace: lrp.Namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: r.createIngressRules(lrp),
		},
	}
}

func (r *RoutingReconciler) createService(lrp *eiriniv1.LRP) (*corev1.Service, error) {
	service := r.generateServiceSpec(lrp)
	if err := ctrl.SetControllerReference(lrp, service, r.scheme); err != nil {
		return nil, err
	}

	return r.clientset.CoreV1().Services(lrp.Namespace).Create(context.Background(), service, metav1.CreateOptions{})
}

func (r *RoutingReconciler) createIngressRules(lrp *eiriniv1.LRP) []v1beta1.IngressRule {
	hostname := fmt.Sprintf("%s.%s.vcap.me", lrp.Spec.AppName, lrp.Namespace)
	port := 8080
	if len(lrp.Spec.Ports) > 0 {
		port = int(lrp.Spec.Ports[0])
	}

	return []v1beta1.IngressRule{
		{
			Host: hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{
						{
							Path: "/",
							Backend: v1beta1.IngressBackend{
								ServiceName: lrp.Spec.GUID,
								ServicePort: intstr.FromInt(port),
							},
						},
					},
				},
			},
		},
	}
}
