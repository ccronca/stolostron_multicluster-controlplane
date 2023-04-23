package klusterlet

import (
	"context"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	operatorv1client "open-cluster-management.io/api/client/operator/clientset/versioned/typed/operator/v1"
	operatorv1informers "open-cluster-management.io/api/client/operator/informers/externalversions/operator/v1"
	workclient "open-cluster-management.io/api/client/work/clientset/versioned"

	"github.com/stolostron/multicluster-controlplane/pkg/controllers/klusterlet/controllers/klusterletcontroller"
	"github.com/stolostron/multicluster-controlplane/pkg/controllers/klusterlet/controllers/statuscontroller"
)

type Klusterlet struct {
	klusterletController factory.Controller
	statusController     factory.Controller
	cleanupController    factory.Controller
}

func (k *Klusterlet) Start(ctx context.Context) {
	go k.klusterletController.Run(ctx, 1)
	go k.cleanupController.Run(ctx, 1)
	go k.statusController.Run(ctx, 1)
}

func NewKlusterlet(
	controlplaneKubeClient kubernetes.Interface,
	controlplaneAPIExtensionClient apiextensionsclient.Interface,
	klusterletClient operatorv1client.KlusterletInterface,
	kubeClient kubernetes.Interface,
	workClient workclient.Interface,
	kubeInformerFactory informers.SharedInformerFactory,
	klusterletInformer operatorv1informers.KlusterletInformer,
) *Klusterlet {
	recorder := events.NewInMemoryRecorder("klusterlet-controller")

	// TODO need go through cleanup controller and do more test
	// TODO enable bootstrap controller?
	// TODO enable sar controller?

	return &Klusterlet{
		klusterletController: klusterletcontroller.NewKlusterletController(
			kubeClient,
			controlplaneKubeClient,
			controlplaneAPIExtensionClient,
			klusterletClient,
			klusterletInformer,
			kubeInformerFactory.Core().V1().Secrets(),
			kubeInformerFactory.Apps().V1().Deployments(),
			workClient.WorkV1().AppliedManifestWorks(),
			recorder,
		),
		cleanupController: klusterletcontroller.NewKlusterletCleanupController(
			kubeClient,
			controlplaneKubeClient,
			controlplaneAPIExtensionClient,
			klusterletClient,
			klusterletInformer,
			kubeInformerFactory.Core().V1().Secrets(),
			kubeInformerFactory.Apps().V1().Deployments(),
			workClient.WorkV1().AppliedManifestWorks(),
			recorder,
		),
		statusController: statuscontroller.NewKlusterletStatusController(
			kubeClient,
			klusterletClient,
			klusterletInformer,
			kubeInformerFactory.Apps().V1().Deployments(),
			recorder,
		),
	}
}
