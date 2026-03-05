//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ../../../fakes/fake_resolver.go . StepResolver
package resolver

import (
	"context"
	"fmt"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	v1listers "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/listers/operators/v1"
	v1alpha1listers "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/listers/operators/v1alpha1"
	controllerbundle "github.com/operator-framework/operator-lifecycle-manager/pkg/controller/bundle"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/cache"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/projection"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorlister"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	BundleLookupConditionPacked v1alpha1.BundleLookupConditionType = "BundleLookupNotPersisted"
	exclusionAnnotation         string                             = "olm.operatorframework.io/exclude-global-namespace-resolution"
)

// init hooks provides the downstream a way to modify the upstream behavior
var initHooks []stepResolverInitHook

type StepResolver interface {
	ResolveSteps(namespace string) ([]*v1alpha1.Step, []v1alpha1.BundleLookup, []*v1alpha1.Subscription, error)
}

type OperatorStepResolver struct {
	subLister              v1alpha1listers.SubscriptionLister
	csvLister              v1alpha1listers.ClusterServiceVersionLister
	ogLister               v1listers.OperatorGroupLister
	client                 versioned.Interface
	globalCatalogNamespace string
	resolver               *Resolver
	log                    logrus.FieldLogger
}

var _ StepResolver = &OperatorStepResolver{}

type catsrcPriorityProvider struct {
	lister v1alpha1listers.CatalogSourceLister
}

func (pp catsrcPriorityProvider) Priority(key cache.SourceKey) int {
	catsrc, err := pp.lister.CatalogSources(key.Namespace).Get(key.Name)
	if err != nil {
		return 0
	}
	return catsrc.Spec.Priority
}

func NewOperatorCacheProvider(lister operatorlister.OperatorLister, client versioned.Interface, sourceProvider cache.SourceProvider, log logrus.FieldLogger) cache.OperatorCacheProvider {
	cacheSourceProvider := &mergedSourceProvider{
		sps: []cache.SourceProvider{
			sourceProvider,
			//SourceProviderFromRegistryClientProvider(provider, log),
			&csvSourceProvider{
				csvLister: lister.OperatorsV1alpha1().ClusterServiceVersionLister(),
				subLister: lister.OperatorsV1alpha1().SubscriptionLister(),
				ogLister:  lister.OperatorsV1().OperatorGroupLister(),
				logger:    log,
				client:    client,
			},
		},
	}
	catSrcPriorityProvider := &catsrcPriorityProvider{lister: lister.OperatorsV1alpha1().CatalogSourceLister()}

	return cache.New(cacheSourceProvider, cache.WithLogger(log), cache.WithSourcePriorityProvider(catSrcPriorityProvider))
}

func NewOperatorStepResolver(lister operatorlister.OperatorLister, client versioned.Interface, globalCatalogNamespace string, opCacheProvider cache.OperatorCacheProvider, log logrus.FieldLogger) *OperatorStepResolver {
	stepResolver := &OperatorStepResolver{
		subLister:              lister.OperatorsV1alpha1().SubscriptionLister(),
		csvLister:              lister.OperatorsV1alpha1().ClusterServiceVersionLister(),
		ogLister:               lister.OperatorsV1().OperatorGroupLister(),
		client:                 client,
		globalCatalogNamespace: globalCatalogNamespace,
		resolver:               NewDefaultResolver(opCacheProvider, log),
		log:                    log,
	}

	// init hooks can be added to the downstream to
	// modify resolver behaviour
	for _, initHook := range initHooks {
		if err := initHook(stepResolver); err != nil {
			panic(err)
		}
	}
	return stepResolver
}

func (r *OperatorStepResolver) ResolveSteps(namespace string) ([]*v1alpha1.Step, []v1alpha1.BundleLookup, []*v1alpha1.Subscription, error) {
	resolutionID := NewResolutionID()
	log := r.log.WithField("resolution_id", resolutionID)
	resolveStart := traceBegin(log, "resolve_steps", kv("namespace", namespace))
	var resolvedSubCount int
	var resolveErr error
	defer func() {
		kvs := []string{kv("namespace", namespace), kv("subscription_count", resolvedSubCount)}
		if resolveErr != nil {
			kvs = append(kvs, kv("error", true))
		}
		traceDone(log, "resolve_steps", resolveStart, kvs...)
	}()

	listSubsStart := traceBegin(log, "list_subscriptions", kv("namespace", namespace))
	subs, err := r.listSubscriptions(namespace)
	if err != nil {
		resolveErr = err
		traceDone(log, "list_subscriptions", listSubsStart, kv("namespace", namespace), kv("error", true))
		return nil, nil, nil, err
	}
	resolvedSubCount = len(subs)
	traceDone(log, "list_subscriptions", listSubsStart, kv("namespace", namespace), kv("subscription_count", len(subs)))

	namespaces := []string{namespace}
	listOGStart := traceBegin(log, "list_operator_groups", kv("namespace", namespace))
	ogs, err := r.ogLister.OperatorGroups(namespace).List(labels.Everything())
	if err != nil {
		resolveErr = err
		traceDone(log, "list_operator_groups", listOGStart, kv("namespace", namespace), kv("error", true))
		return nil, nil, nil, fmt.Errorf("listing operatorgroups in namespace %s: %s", namespace, err)
	}
	traceDone(log, "list_operator_groups", listOGStart, kv("namespace", namespace), kv("count", len(ogs)))
	if len(ogs) != 1 {
		resolveErr = fmt.Errorf("expected 1 OperatorGroup in the namespace, found %d", len(ogs))
		return nil, nil, nil, resolveErr
	}
	og := ogs[0]
	if val, ok := og.Annotations[exclusionAnnotation]; ok && val == "true" {
		// Exclusion specified
		// Ignore the globalNamespace for the purposes of resolution in this namespace
		log.Printf("excluding global catalogs from resolution in namespace %s", namespace)
	} else {
		namespaces = append(namespaces, r.globalCatalogNamespace)
	}
	coreResolveStart := traceBegin(log, "core_resolve", kv("namespace", namespace), kv("subscription_count", len(subs)), kv("namespace_count", len(namespaces)))
	operators, err := r.resolver.WithLog(log).Resolve(namespaces, subs)
	if err != nil {
		resolveErr = err
		traceDone(log, "core_resolve", coreResolveStart, kv("namespace", namespace), kv("error", true))
		return nil, nil, nil, err
	}
	traceDone(log, "core_resolve", coreResolveStart, kv("namespace", namespace), kv("subscription_count", len(subs)), kv("operator_count", len(operators)))

	// if there's no error, we were able to satisfy all constraints in the subscription set, so we calculate what
	// changes to persist to the cluster and write them out as `steps`
	generateStepsStart := traceBegin(log, "generate_steps", kv("namespace", namespace), kv("operator_count", len(operators)))
	steps := []*v1alpha1.Step{}
	updatedSubs := []*v1alpha1.Subscription{}
	bundleLookups := []v1alpha1.BundleLookup{}
	for _, op := range operators {
		// Find any existing subscriptions that resolve to this operator.
		existingSubscriptions := make(map[*v1alpha1.Subscription]bool)
		sourceInfo := *op.SourceInfo
		for _, sub := range subs {
			if sub.Spec.Package != sourceInfo.Package {
				continue
			}
			if sub.Spec.Channel != "" && sub.Spec.Channel != sourceInfo.Channel {
				continue
			}
			subCatalogKey := cache.SourceKey{
				Name:      sub.Spec.CatalogSource,
				Namespace: sub.Spec.CatalogSourceNamespace,
			}
			if !subCatalogKey.Empty() && !subCatalogKey.Equal(sourceInfo.Catalog) {
				continue
			}
			alreadyExists, err := r.hasExistingCurrentCSV(sub)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("unable to determine whether subscription %s has a preexisting CSV", sub.GetName())
			}
			existingSubscriptions[sub] = alreadyExists
		}

		if len(existingSubscriptions) > 0 {
			upToDate := true
			for sub, exists := range existingSubscriptions {
				if !exists || sub.Status.CurrentCSV != op.Name {
					upToDate = false
				}
			}
			// all matching subscriptions are up to date
			if upToDate {
				continue
			}
		}

		// add steps for any new bundle
		if op.Bundle != nil {
			bundleSteps, err := NewStepsFromBundle(op.Bundle, namespace, op.Replaces, op.SourceInfo.Catalog.Name, op.SourceInfo.Catalog.Namespace)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to turn bundle into steps: %s", err.Error())
			}
			steps = append(steps, bundleSteps...)
		} else {
			lookup := v1alpha1.BundleLookup{
				Path:       op.BundlePath,
				Identifier: op.Name,
				Replaces:   op.Replaces,
				CatalogSourceRef: &corev1.ObjectReference{
					Namespace: op.SourceInfo.Catalog.Namespace,
					Name:      op.SourceInfo.Catalog.Name,
				},
				Conditions: []v1alpha1.BundleLookupCondition{
					{
						Type:    BundleLookupConditionPacked,
						Status:  corev1.ConditionTrue,
						Reason:  controllerbundle.NotUnpackedReason,
						Message: controllerbundle.NotUnpackedMessage,
					},
					{
						Type:    v1alpha1.BundleLookupPending,
						Status:  corev1.ConditionTrue,
						Reason:  controllerbundle.JobNotStartedReason,
						Message: controllerbundle.JobNotStartedMessage,
					},
				},
			}
			anno, err := projection.PropertiesAnnotationFromPropertyList(op.Properties)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to serialize operator properties for %q: %w", op.Name, err)
			}
			lookup.Properties = anno
			bundleLookups = append(bundleLookups, lookup)
		}

		if len(existingSubscriptions) == 0 {
			// explicitly track the resolved CSV as the starting CSV on the resolved subscriptions
			op.SourceInfo.StartingCSV = op.Name
			subStep, err := NewSubscriptionStepResource(namespace, *op.SourceInfo)
			if err != nil {
				return nil, nil, nil, err
			}
			steps = append(steps, &v1alpha1.Step{
				Resolving: op.Name,
				Resource:  subStep,
				Status:    v1alpha1.StepStatusUnknown,
			})
		}

		// add steps for subscriptions for bundles that were added through resolution
		for sub := range existingSubscriptions {
			if sub.Status.CurrentCSV == op.Name {
				continue
			}
			// update existing subscription status
			sub.Status.CurrentCSV = op.Name
			updatedSubs = append(updatedSubs, sub)
		}
	}

	// Order Steps
	steps = v1alpha1.OrderSteps(steps)
	traceDone(log, "generate_steps", generateStepsStart, kv("namespace", namespace), kv("step_count", len(steps)), kv("bundle_lookup_count", len(bundleLookups)))
	return steps, bundleLookups, updatedSubs, nil
}

func (r *OperatorStepResolver) hasExistingCurrentCSV(sub *v1alpha1.Subscription) (bool, error) {
	if sub.Status.CurrentCSV == "" {
		return false, nil
	}
	_, err := r.csvLister.ClusterServiceVersions(sub.GetNamespace()).Get(sub.Status.CurrentCSV)
	if err == nil {
		return true, nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}
	return false, err // Can't answer this question right now.
}

func (r *OperatorStepResolver) listSubscriptions(namespace string) ([]*v1alpha1.Subscription, error) {
	list, err := r.client.OperatorsV1alpha1().Subscriptions(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var subs []*v1alpha1.Subscription
	for i := range list.Items {
		subs = append(subs, &list.Items[i])
	}

	return subs, nil
}

type mergedSourceProvider struct {
	sps    []cache.SourceProvider
	logger logrus.StdLogger
}

func (msp *mergedSourceProvider) Sources(namespaces ...string) map[cache.SourceKey]cache.Source {
	result := make(map[cache.SourceKey]cache.Source)
	for _, sp := range msp.sps {
		for key, source := range sp.Sources(namespaces...) {
			if _, ok := result[key]; ok {
				msp.logger.Printf("warning: duplicate sourcekey: %q\n", key)
			}
			result[key] = source
		}
	}
	return result
}
