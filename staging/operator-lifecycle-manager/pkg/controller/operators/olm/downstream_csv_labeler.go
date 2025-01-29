package olm

import (
	"context"
	"fmt"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/olm/plugins"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CsvLabelerPluginID plugins.PluginID = "csv-labeler-plugin"
const labelSyncerLabelKey = ""

func NewCSVLabelSyncerLabeler(client operatorclient.ClientInterface, logger *logrus.Logger) *CSVLabelSyncerLabeler {
	return &CSVLabelSyncerLabeler{
		client: client,
		logger: logger,
	}
}

type CSVLabelSyncerLabeler struct {
	client operatorclient.ClientInterface
	logger *logrus.Logger
}

func (c *CSVLabelSyncerLabeler) OnAddOrUpdate(csv *v1alpha1.ClusterServiceVersion) error {
	// ignore copied csvs
	if csv.IsCopied() {
		return nil
	}

	// ignore csv updates
	if csv.Status.LastTransitionTime != nil {
		return nil
	}

	namespace, err := c.client.KubernetesInterface().CoreV1().Namespaces().Get(context.Background(), csv.GetNamespace(), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting csv namespace (%s) for label sync'er labeling", csv.GetNamespace())
	}

	// add label sync'er label if it does not exist
	if _, ok := namespace.Labels[labelSyncerLabelKey]; !ok {
		nsCopy := namespace.DeepCopy()
		nsCopy.Labels[labelSyncerLabelKey] = "true"
		if _, err := c.client.KubernetesInterface().CoreV1().Namespaces().Update(context.Background(), namespace, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("error updating csv namespace (%s) with label sync'er label", nsCopy.GetNamespace())
		}

		if c.logger != nil {
			c.logger.Printf("[CSV LABEL] applied %s=true label to namespace %s", labelSyncerLabelKey, nsCopy.GetNamespace())
		}
	}

	return nil
}

func (c *CSVLabelSyncerLabeler) OnDelete(_ *v1alpha1.ClusterServiceVersion) error {
	return nil
}
