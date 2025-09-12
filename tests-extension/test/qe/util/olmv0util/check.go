package olmv0util

import (
	"fmt"
	"sync"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
)

type checkDescription struct {
	method          string
	executor        bool
	inlineNamespace bool
	expectAction    bool
	expectContent   string
	expected        bool // Whether the check should pass (true) or fail (false)
	resource        []string
}

// the method is to make newCheck object.
// the method parameter is expect, it will check something is expceted or not
// the method parameter is present, it will check something exists or not
// the executor is exutil.AsAdmin, it will exectue oc with Admin
// the executor is AsUser, it will exectue oc with User
// the inlineNamespace is exutil.WithoutNamespace, it will execute oc with exutil.WithoutNamespace()
// the inlineNamespace is WithNamespace, it will execute oc with WithNamespace()
// the expectAction take effective when method is expect, if it is Contain, it will check if the strings Contain substring with expectContent parameter
//
//	if it is Compare, it will check the strings is samme with expectContent parameter
//
// the expectContent is the content we expected
// the expect is ok, Contain or Compare result is OK for method == expect, no error raise. if not OK, error raise
// the expect is nok, Contain or Compare result is NOK for method == expect, no error raise. if OK, error raise
// the expect is ok, resource existing is OK for method == present, no error raise. if resource not existing, error raise
// the expect is nok, resource not existing is OK for method == present, no error raise. if resource existing, error raise
func NewCheck(method string, executor bool, inlineNamespace bool, expectAction bool,
	expectContent string, expect bool, resource []string) checkDescription {
	return checkDescription{
		method:          method,
		executor:        executor,
		inlineNamespace: inlineNamespace,
		expectAction:    expectAction,
		expectContent:   expectContent,
		expected:        expect,
		resource:        resource,
	}
}

// the method is to check the resource per definition of the above described newCheck.
func (ck checkDescription) Check(oc *exutil.CLI) {
	switch ck.method {
	case "present":
		ok := IsPresentResource(oc, ck.executor, ck.inlineNamespace, ck.expectAction, ck.resource...)
		o.Expect(ok).To(o.BeTrue())
	case "expect":
		err := expectedResource(oc, ck.executor, ck.inlineNamespace, ck.expectAction, ck.expectContent, ck.expected, ck.resource...)
		if err != nil {
			ns := oc.Namespace()
			for i, v := range ck.resource {
				if v == "-n" && i+1 < len(ck.resource) {
					ns = ck.resource[i+1]
					break
				}
			}
			GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "pod", "-n", "openshift-marketplace")
			GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "operatorgroup", "-n", ns, "-o", "yaml")
			GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "catalogsource", "-n", ns, "-o", "yaml")
			GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "subscription", "-n", ns, "-o", "yaml")
			GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", "-n", ns)
			GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "csv", "-n", ns)
			GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "pods", "-n", ns)
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("expected content %s not found by %v", ck.expectContent, ck.resource))
	default:
		err := fmt.Errorf("unknown method")
		o.Expect(err).NotTo(o.HaveOccurred())
	}
}

// the method is to check the resource, but not assert it which is diffrence with the method check().
func (ck checkDescription) CheckWithoutAssert(oc *exutil.CLI) error {
	switch ck.method {
	case "present":
		ok := IsPresentResource(oc, ck.executor, ck.inlineNamespace, ck.expectAction, ck.resource...)
		if ok {
			return nil
		}
		return fmt.Errorf("it is not epxected")
	case "expect":
		return expectedResource(oc, ck.executor, ck.inlineNamespace, ck.expectAction, ck.expectContent, ck.expected, ck.resource...)
	default:
		return fmt.Errorf("unknown method")
	}
}

// it is the check list so that all the check are done in parallel.
type CheckList []checkDescription

// the method is to add one check
func (cl CheckList) Add(ck checkDescription) CheckList {
	cl = append(cl, ck)
	return cl
}

// the method is to make check list empty.
func (cl CheckList) Empty() CheckList {
	cl = cl[0:0]
	return cl
}

// the method is to execute all the check in parallel.
func (cl CheckList) Check(oc *exutil.CLI) {
	var wg sync.WaitGroup
	for _, ck := range cl {
		wg.Add(1)
		go func(ck checkDescription) {
			defer g.GinkgoRecover()
			defer wg.Done()
			ck.Check(oc)
		}(ck)
	}
	wg.Wait()
}
