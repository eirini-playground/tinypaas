package main

import (
	"log"
	"os"

	eirinischeme "code.cloudfoundry.org/eirini/pkg/generated/clientset/versioned/scheme"
	"code.cloudfoundry.org/lager"
	"github.com/go-logr/logr"
	"github.com/jimmykarily/tinypaas/image"
	kpackv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackscheme "github.com/pivotal/kpack/pkg/client/clientset/versioned/scheme"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func main() {
	err := kscheme.AddToScheme(kpackscheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}

	err = eirinischeme.AddToScheme(kpackscheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatal(err)
	}

	runtimeClient, err := runtimeclient.New(kubeConfig, runtimeclient.Options{Scheme: kpackscheme.Scheme})
	if err != nil {
		log.Fatalf("Failed to create k8s runtime client: %s", err)
	}

	_, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("Failed to create k8s clientset: %s", err)
	}

	logger := lager.NewLogger("image-controller")
	logger.RegisterSink(lager.NewPrettySink(os.Stdout, lager.DEBUG))

	managerOptions := manager.Options{
		MetricsBindAddress: "0",
		Scheme:             kpackscheme.Scheme,
		Logger:             NewLagerLogr(logger),
		LeaderElection:     true,
		LeaderElectionID:   "image-controller-leader",
	}

	mgr, err := manager.New(kubeConfig, managerOptions)
	if err != nil {
		log.Fatalf("Failed to create k8s controller runtime manager: %s", err)
	}

	imageReconciler := image.NewReconciler(logger, runtimeClient)

	err = builder.
		ControllerManagedBy(mgr).
		For(&kpackv1alpha1.Image{}).
		Complete(imageReconciler)

	if err != nil {
		log.Fatalf("Failed to build Image reconciler: %s", err)
	}

	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		log.Fatalf("Failed to start manager: %s", err)
	}
}

type LagerLogr struct {
	logger lager.Logger
}

func (l LagerLogr) Info(msg string, kvs ...interface{}) {
	l.logger.Info(msg, toLagerData(kvs))
}

func (l LagerLogr) Enabled() bool {
	return true
}

func NewLagerLogr(logger lager.Logger) logr.Logger {
	return LagerLogr{
		logger: logger,
	}
}

func (l LagerLogr) Error(err error, msg string, kvs ...interface{}) {
	l.logger.Error(msg, err, toLagerData(kvs))
}

func (l LagerLogr) V(level int) logr.Logger {
	return l
}

func (l LagerLogr) WithValues(kvs ...interface{}) logr.Logger {
	return l
}

func (l LagerLogr) WithName(name string) logr.Logger {
	return l
}

func toLagerData(kvs ...interface{}) lager.Data {
	return lager.Data{"data": kvs}
}
