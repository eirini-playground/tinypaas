package main

import (
	"os"

	eiriniv1 "code.cloudfoundry.org/eirini/pkg/apis/eirini/v1"
	eirinischeme "code.cloudfoundry.org/eirini/pkg/generated/clientset/versioned/scheme"
	"code.cloudfoundry.org/lager"
	"github.com/jimmykarily/tinypaas/routing/k8s"
	"github.com/jimmykarily/tinypaas/util"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func main() {
	if err := kscheme.AddToScheme(eirinischeme.Scheme); err != nil {
		util.Exitf("failed to add the k8s scheme to the LRP CRD scheme: %v", err)
	}

	kubeConfig, err := rest.InClusterConfig()
	util.ExitfIfError(err, "Failed to build kubeconfig")

	controllerClient, err := runtimeclient.New(kubeConfig, runtimeclient.Options{Scheme: eirinischeme.Scheme})
	util.ExitfIfError(err, "Failed to create k8s runtime client")

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	util.ExitfIfError(err, "Failed to create k8s clientset")

	logger := lager.NewLogger("eirini-routing")
	logger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))

	managerOptions := manager.Options{
		// do not serve prometheus metrics; disabled because port clashes during integration tests
		MetricsBindAddress: "0",
		Scheme:             eirinischeme.Scheme,
		Logger:             util.NewLagerLogr(logger),
		LeaderElection:     true,
		LeaderElectionID:   "eirini-routing-leader",
	}

	mgr, err := manager.New(kubeConfig, managerOptions)
	util.ExitfIfError(err, "Failed to create k8s controller runtime manager")

	lrpReconciler := createRoutingReconciler(logger, controllerClient, clientset, mgr.GetScheme())

	err = builder.
		ControllerManagedBy(mgr).
		For(&eiriniv1.LRP{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(lrpReconciler)
	util.ExitfIfError(err, "Failed to build LRP reconciler")

	err = mgr.Start(ctrl.SetupSignalHandler())
	util.ExitfIfError(err, "Failed to start manager")
}

func createRoutingReconciler(
	logger lager.Logger,
	controllerClient runtimeclient.Client,
	clientset kubernetes.Interface,
	scheme *runtime.Scheme,
) *k8s.RoutingReconciler {
	return k8s.NewRoutingReconciler(
		logger,
		controllerClient,
		clientset,
		scheme)
}
