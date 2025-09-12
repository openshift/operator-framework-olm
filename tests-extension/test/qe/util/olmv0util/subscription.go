// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains Subscription management utilities for installing
// and managing operator subscriptions in OpenShift/Kubernetes environments
package olmv0util

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

type SubscriptionDescription struct {
	SubName                string `json:"name"`
	Namespace              string `json:"namespace"`
	Channel                string `json:"channel"`
	IpApproval             string `json:"installPlanApproval"`
	OperatorPackage        string `json:"spec.name"`
	CatalogSourceName      string `json:"source"`
	CatalogSourceNamespace string `json:"sourceNamespace"`
	StartingCSV            string `json:"startingCSV,omitempty"`
	ConfigMapRef           string `json:"configMapRef,omitempty"`
	SecretRef              string `json:"secretRef,omitempty"`
	CurrentCSV             string
	InstalledCSV           string
	Template               string
	SingleNamespace        bool
	IpCsv                  string
	ClusterType            string
}

// the method is to create sub, and save the sub resrouce into dr. and more create csv possible depending on sub.IpApproval
// if sub.IpApproval is Automatic, it will wait the sub's state become AtLatestKnown and get installed csv as sub.InstalledCSV, and save csv into dr
// if sub.IpApproval is not Automatic, it will just wait sub's state become UpgradePending
func (sub *SubscriptionDescription) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	// for most operator subscription failure, the reason is that there is a left cluster-scoped CSV.
	// I'd like to print all CSV before create it.
	// allCSVs, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "--all-namespaces").Output()
	// if err != nil {
	// 	e2e.Failf("!!! Couldn't get all CSVs:%v\n", err)
	// }
	// e2e.Logf("!!! Get all CSVs in this cluster:\n%s\n", allCSVs)

	sub.CreateWithoutCheck(oc, itName, dr)
	if strings.Compare(sub.IpApproval, "Automatic") == 0 {
		sub.FindInstalledCSVWithSkip(oc, itName, dr, true)
	} else {
		NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "UpgradePending", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.state}"}).Check(oc)
	}
}

// It's for the manual subscription to get its latest status, such as, the installedCSV.
func (sub *SubscriptionDescription) Update(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	installedCSV := GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installedCSV}")
	o.Expect(installedCSV).NotTo(o.BeEmpty())
	if strings.Compare(sub.InstalledCSV, installedCSV) != 0 {
		sub.InstalledCSV = installedCSV
		dr.GetIr(itName).Add(NewResource(oc, "csv", sub.InstalledCSV, exutil.RequireNS, sub.Namespace))
	}
	e2e.Logf("updating the subscription to get the latest installedCSV: %s", sub.InstalledCSV)
}

// the method is to just create sub, and save it to dr, do not check its state.
// Note that, this func doesn't get the installedCSV, this may lead to your operator CSV won't be deleted when calling sub.deleteCSV()
func (sub *SubscriptionDescription) CreateWithoutCheck(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	//isAutomatic := strings.Compare(sub.IpApproval, "Automatic") == 0

	//startingCSV is not necessary. And, if there are multi same package from different CatalogSource, it will lead to error.
	//if strings.Compare(sub.currentCSV, "") == 0 {
	//	sub.currentCSV = GetResource(oc, AsAdmin, WithoutNamespace, "packagemanifest", sub.OperatorPackage, fmt.Sprintf("-o=jsonpath={.status.channels[?(@.name==\"%s\")].currentCSV}", sub.Channel))
	//	o.Expect(sub.currentCSV).NotTo(o.BeEmpty())
	//}

	//if isAutomatic {
	//	sub.StartingCSV = sub.currentCSV
	//} else {
	//	o.Expect(sub.StartingCSV).NotTo(o.BeEmpty())
	//}

	// for most operator subscription failure, the reason is that there is a left cluster-scoped CSV.
	// I'd like to print all CSV before create it.
	// It prints many lines which descrease the exact match for RP, and increase log size.
	// So, change it to one line with necessary information csv name and namespace.
	allCSVs, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "--all-namespaces", "-o=jsonpath={range .items[*]}{@.metadata.name}{\",\"}{@.metadata.namespace}{\":\"}{end}").Output()
	if err != nil {
		if strings.Contains(allCSVs, "unexpected EOF") || strings.Contains(err.Error(), "status 1") {
			g.Skip(fmt.Sprintf("skip case with %v", allCSVs+err.Error()))
		}
		e2e.Failf("!!! Couldn't get all CSVs:%v\n", err)
	}
	csvMap := make(map[string][]string)
	csvList := strings.Split(allCSVs, ":")
	for _, csv := range csvList {
		if strings.Compare(csv, "") == 0 {
			continue
		}
		name := strings.Split(csv, ",")[0]
		ns := strings.Split(csv, ",")[1]
		val, ok := csvMap[name]
		if ok {
			if strings.HasPrefix(ns, "openshift-") {
				alreadyOpenshiftDefaultNS := false
				for _, v := range val {
					if strings.Contains(v, "openshift-") {
						alreadyOpenshiftDefaultNS = true // normally one default operator exists in all openshift- ns, like elasticsearch-operator
						// only add one openshift- ns to indicate. to save log size and line size. Or else one line
						// will be greater than 3k
						break
					}
				}
				if !alreadyOpenshiftDefaultNS {
					val = append(val, ns)
					csvMap[name] = val
				}
			} else {
				val = append(val, ns)
				csvMap[name] = val
			}
		} else {
			nsSlice := make([]string, 20)
			nsSlice[1] = ns
			csvMap[name] = nsSlice
		}
	}
	for name, ns := range csvMap {
		e2e.Logf("getting csv is %v, the related NS is %v", name, ns)
	}

	e2e.Logf("create sub %s", sub.SubName)
	applyFn := ApplyResourceFromTemplate
	if strings.Compare(sub.ClusterType, "microshift") == 0 {
		applyFn = ApplyResourceFromTemplateOnMicroshift
	}
	err = applyFn(oc, "--ignore-unknown-parameters=true", "-f", sub.Template, "-p", "SUBNAME="+sub.SubName, "SUBNAMESPACE="+sub.Namespace, "CHANNEL="+sub.Channel,
		"APPROVAL="+sub.IpApproval, "OPERATORNAME="+sub.OperatorPackage, "SOURCENAME="+sub.CatalogSourceName, "SOURCENAMESPACE="+sub.CatalogSourceNamespace,
		"STARTINGCSV="+sub.StartingCSV, "CONFIGMAPREF="+sub.ConfigMapRef, "SECRETREF="+sub.SecretRef)

	o.Expect(err).NotTo(o.HaveOccurred())
	dr.GetIr(itName).Add(NewResource(oc, "sub", sub.SubName, exutil.RequireNS, sub.Namespace))
}
func (sub *SubscriptionDescription) CreateWithoutCheckNoPrint(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	e2e.Logf("create sub %s", sub.SubName)
	applyFn := ApplyResourceFromTemplate
	if strings.Compare(sub.ClusterType, "microshift") == 0 {
		applyFn = ApplyResourceFromTemplateOnMicroshift
	}
	err := applyFn(oc, "--ignore-unknown-parameters=true", "-f", sub.Template, "-p", "SUBNAME="+sub.SubName, "SUBNAMESPACE="+sub.Namespace, "CHANNEL="+sub.Channel,
		"APPROVAL="+sub.IpApproval, "OPERATORNAME="+sub.OperatorPackage, "SOURCENAME="+sub.CatalogSourceName, "SOURCENAMESPACE="+sub.CatalogSourceNamespace,
		"STARTINGCSV="+sub.StartingCSV, "CONFIGMAPREF="+sub.ConfigMapRef, "SECRETREF="+sub.SecretRef)

	o.Expect(err).NotTo(o.HaveOccurred())
	dr.GetIr(itName).Add(NewResource(oc, "sub", sub.SubName, exutil.RequireNS, sub.Namespace))
}

// the method is to check if the sub's state is AtLatestKnown.
// if it is AtLatestKnown, get installed csv from sub and save it to dr.
// if it is not AtLatestKnown, raise error.
func (sub *SubscriptionDescription) FindInstalledCSV(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	sub.FindInstalledCSVWithSkip(oc, itName, dr, false)
}

func (sub *SubscriptionDescription) FindInstalledCSVWithSkip(oc *exutil.CLI, itName string, dr DescriberResrouce, skip bool) {
	err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
		state, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.state}").Output()
		if strings.Compare(state, "AtLatestKnown") == 0 {
			return true, nil
		}
		e2e.Logf("sub %s state is %s, not AtLatestKnown", sub.SubName, state)
		return false, nil
	})
	if err != nil {
		message, _ := oc.AsAdmin().WithoutNamespace().Run("describe").Args("sub", sub.SubName, "-n", sub.Namespace).Output()
		e2e.Logf("Subscription describe output: %s", message)
		if sub.AssertToSkipSpecificMessage(message) && skip {
			g.Skip(fmt.Sprintf("the case skip without issue and impacted by others: %s", message))
		}
		message, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace,
			"-o=jsonpath={.status.conditions[?(@.type==\"ResolutionFailed\")].message}").Output()
		if sub.AssertToSkipSpecificMessage(message) && skip {
			g.Skip(fmt.Sprintf("the case skip without issue and impacted by others: %s", message))
		}
		message, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("installplan", "-n", sub.Namespace, "-o=jsonpath-as-json={..status}").Output()
		e2e.Logf("InstallPlan status: %s", message)
		message, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", sub.CatalogSourceNamespace).Output()
		e2e.Logf("Pods in catalog source namespace: %s", message)
		message, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("pod", "-n", sub.Namespace).Output()
		e2e.Logf("Pods in subscription namespace: %s", message)
		message, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("event", "-n", sub.Namespace).Output()
		e2e.Logf("Events in subscription namespace: %s", message)
	}
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("sub %s stat is not AtLatestKnown", sub.SubName))

	installedCSV := GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installedCSV}")
	o.Expect(installedCSV).NotTo(o.BeEmpty())
	if strings.Compare(sub.InstalledCSV, installedCSV) != 0 {
		sub.InstalledCSV = installedCSV
		dr.GetIr(itName).Add(NewResource(oc, "csv", sub.InstalledCSV, exutil.RequireNS, sub.Namespace))
	}
	e2e.Logf("the installed CSV name is %s", sub.InstalledCSV)
}

func (sub *SubscriptionDescription) AssertToSkipSpecificMessage(message string) bool {
	specificMessages := []string{
		"subscription sub-learn-46964 requires @existing/openshift-operators//learn-operator.v0.0.3",
		"error using catalogsource openshift-marketplace/qe-app-registry",
		"failed to list bundles: rpc error: code = Unavailable desc = connection error",
		"Unable to connect to the server",
	}
	for _, specificMessage := range specificMessages {
		if strings.Contains(message, specificMessage) {
			return true
		}
	}
	return false

}

// the method is to check if the cv parameter is same to the installed csv.
// if not same, raise error.
// if same, nothong happen.
func (sub *SubscriptionDescription) ExpectCSV(oc *exutil.CLI, itName string, dr DescriberResrouce, cv string) {
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 480*time.Second, false, func(ctx context.Context) (bool, error) {
		sub.FindInstalledCSV(oc, itName, dr)
		if strings.Compare(sub.InstalledCSV, cv) == 0 {
			return true, nil
		}
		return false, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("expected csv %s not found", cv))
}

// the method is to approve the install plan when you create sub with sub.IpApproval != Automatic
// normally firstly call sub.create(), then call this method sub.approve. it is used to operator upgrade case.
func (sub *SubscriptionDescription) Approve(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	err := wait.PollUntilContextTimeout(context.TODO(), 6*time.Second, 360*time.Second, false, func(ctx context.Context) (bool, error) {
		for strings.Compare(sub.InstalledCSV, "") == 0 {
			state := GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.state}")
			if strings.Compare(state, "AtLatestKnown") == 0 {
				sub.InstalledCSV = GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installedCSV}")
				dr.GetIr(itName).Add(NewResource(oc, "csv", sub.InstalledCSV, exutil.RequireNS, sub.Namespace))
				e2e.Logf("it is already done, and the installed CSV name is %s", sub.InstalledCSV)
				continue
			}

			ipCsv := GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installplan.name}{\" \"}{.status.currentCSV}")
			sub.IpCsv = ipCsv + "##" + sub.IpCsv
			installPlan := strings.Fields(ipCsv)[0]
			o.Expect(installPlan).NotTo(o.BeEmpty())
			e2e.Logf("try to approve installPlan %s", installPlan)
			PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", installPlan, "-n", sub.Namespace, "--type", "merge", "-p", "{\"spec\": {\"approved\": true}}")
			err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 70*time.Second, false, func(ctx context.Context) (bool, error) {
				err := NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "Complete", exutil.Ok, []string{"installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)
				if err != nil {
					e2e.Logf("the get error is %v, and try next", err)
					return false, nil
				}
				return true, nil
			})
			exutil.AssertWaitPollNoErr(err, fmt.Sprintf("installPlan %s is not Complete", installPlan))
		}
		return true, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("not found installed csv for %s", sub.SubName))
}

// The user can approve the specific InstallPlan:
// NAME            CSV                   APPROVAL   APPROVED
// install-vmwlk   etcdoperator.v0.9.4   Manual     false
// install-xqgtx   etcdoperator.v0.9.2   Manual     true
// sub.approveSpecificIP(oc, itName, dr, "etcdoperator.v0.9.2", "Complete") approve this "etcdoperator.v0.9.2" InstallPlan only
func (sub *SubscriptionDescription) ApproveSpecificIP(oc *exutil.CLI, itName string, dr DescriberResrouce, csvName string, phase string) {
	// fix https://github.com/openshift/openshift-tests-private/issues/735
	var state string
	if err := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 600*time.Second, false, func(ctx context.Context) (bool, error) {
		state = GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.state}")
		if strings.Compare(state, "UpgradePending") == 0 {
			return true, nil
		}
		return false, nil
	}); err != nil {
		e2e.Logf("Failed to wait for UpgradePending state: %v", err)
	}
	if strings.Compare(state, "UpgradePending") == 0 {
		e2e.Logf("--> The expected CSV: %s", csvName)
		err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 90*time.Second, false, func(ctx context.Context) (bool, error) {
			ipCsv := GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installplan.name}{\" \"}{.status.currentCSV}")
			if strings.Contains(ipCsv, csvName) {
				installPlan := strings.Fields(ipCsv)[0]
				if len(installPlan) == 0 {
					return false, fmt.Errorf("installPlan is empty")
				}
				e2e.Logf("---> Get the pending InstallPlan %s", installPlan)
				PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "installplan", installPlan, "-n", sub.Namespace, "--type", "merge", "-p", "{\"spec\": {\"approved\": true}}")
				err := wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 70*time.Second, false, func(ctx context.Context) (bool, error) {
					err := NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, phase, exutil.Ok, []string{"installplan", installPlan, "-n", sub.Namespace, "-o=jsonpath={.status.phase}"}).CheckWithoutAssert(oc)
					if err != nil {
						return false, nil
					}
					return true, nil
				})
				// break the wait loop and return an error
				if err != nil {
					return true, fmt.Errorf("installPlan %s is not %s", installPlan, phase)
				}
				return true, nil
			} else {
				e2e.Logf("--> Not found the expected CSV(%s), the current IP:%s", csvName, ipCsv)
				return false, nil
			}
		})
		if err != nil && strings.Contains(err.Error(), "installPlan") {
			e2e.Failf("InstallPlan error: %s", err.Error())
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("--> Not found the expected CSV: %s", csvName))
	} else {
		CSVs := GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installedCSV}{\" \"}{.status.currentCSV}")
		e2e.Logf("---> No need any apporval operation, the InstalledCSV and currentCSV are the same: %s", CSVs)
	}
}

// the method is to construct one csv object.
func (sub *SubscriptionDescription) GetCSV() CsvDescription {
	e2e.Logf("csv is %s, namespace is %s", sub.InstalledCSV, sub.Namespace)
	return CsvDescription{sub.InstalledCSV, sub.Namespace}
}

// get the reference InstallPlan
func (sub *SubscriptionDescription) GetIP(oc *exutil.CLI) string {
	var installPlan string
	waitErr := wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
		var err error
		installPlan, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.installPlanRef.name}").Output()
		if strings.Compare(installPlan, "") == 0 || err != nil {
			return false, nil
		}
		return true, nil
	})
	exutil.AssertWaitPollNoErr(waitErr, fmt.Sprintf("sub %s has no installplan", sub.SubName))
	o.Expect(installPlan).NotTo(o.BeEmpty())
	return installPlan
}

// the method is to get the CR version from alm-examples of csv if it exists
func (sub *SubscriptionDescription) GetInstanceVersion(oc *exutil.CLI) string {
	version := ""
	output := strings.Split(GetResource(oc, exutil.AsUser, exutil.WithoutNamespace, "csv", sub.InstalledCSV, "-n", sub.Namespace, "-o=jsonpath={.metadata.annotations.alm-examples}"), "\n")
	for _, line := range output {
		if strings.Contains(line, "\"version\"") {
			version = strings.Trim(strings.Fields(strings.TrimSpace(line))[1], "\"")
			break
		}
	}
	o.Expect(version).NotTo(o.BeEmpty())
	e2e.Logf("sub cr version is %s", version)
	return version
}

// the method is obsolete
func (sub *SubscriptionDescription) CreateInstance(oc *exutil.CLI, instance string) {
	path := filepath.Join(e2e.TestContext.OutputDir, sub.Namespace+"-"+"instance.json")
	err := os.WriteFile(path, []byte(instance), 0644)
	o.Expect(err).NotTo(o.HaveOccurred())
	err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-n", sub.Namespace, "-f", path).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
}

// the method is to delete sub which is saved when calling sub.create() or sub.createWithoutCheck()
func (sub *SubscriptionDescription) Delete(itName string, dr DescriberResrouce) {
	e2e.Logf("remove sub %s, ns is %s", sub.SubName, sub.Namespace)
	dr.GetIr(itName).Remove(sub.SubName, "sub", sub.Namespace)
}
func (sub *SubscriptionDescription) DeleteCSV(itName string, dr DescriberResrouce) {
	e2e.Logf("remove csv %s, ns is %s, the subscription name is: %s", sub.InstalledCSV, sub.Namespace, sub.SubName)
	dr.GetIr(itName).Remove(sub.InstalledCSV, "csv", sub.Namespace)
}

// the method is to patch sub object
func (sub *SubscriptionDescription) Patch(oc *exutil.CLI, patch string) {
	PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "sub", sub.SubName, "-n", sub.Namespace, "--type", "merge", "-p", patch)
}

type SubscriptionDescriptionProxy struct {
	SubscriptionDescription
	HttpProxy  string
	HttpsProxy string
	NoProxy    string
}

// the method is to just create sub with proxy, and save it to dr, do not check its state.
func (sub *SubscriptionDescriptionProxy) CreateWithoutCheck(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	e2e.Logf("install subscriptionDescriptionProxy")
	err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", sub.Template, "-p", "SUBNAME="+sub.SubName, "SUBNAMESPACE="+sub.Namespace, "CHANNEL="+sub.Channel,
		"APPROVAL="+sub.IpApproval, "OPERATORNAME="+sub.OperatorPackage, "SOURCENAME="+sub.CatalogSourceName, "SOURCENAMESPACE="+sub.CatalogSourceNamespace, "STARTINGCSV="+sub.StartingCSV,
		"SUBHTTPPROXY="+sub.HttpProxy, "SUBHTTPSPROXY="+sub.HttpsProxy, "SUBNOPROXY="+sub.NoProxy)

	o.Expect(err).NotTo(o.HaveOccurred())
	dr.GetIr(itName).Add(NewResource(oc, "sub", sub.SubName, exutil.RequireNS, sub.Namespace))
	e2e.Logf("install subscriptionDescriptionProxy %s SUCCESS", sub.SubName)
}

func (sub *SubscriptionDescriptionProxy) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	sub.CreateWithoutCheck(oc, itName, dr)
	if strings.Compare(sub.IpApproval, "Automatic") == 0 {
		sub.FindInstalledCSVWithSkip(oc, itName, dr, true)
	} else {
		NewCheck("expect", exutil.AsAdmin, exutil.WithoutNamespace, exutil.Compare, "UpgradePending", exutil.Ok, []string{"sub", sub.SubName, "-n", sub.Namespace, "-o=jsonpath={.status.state}"}).Check(oc)
	}
}
