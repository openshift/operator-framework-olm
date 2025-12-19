package olmv0util

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"path/filepath"
	"strconv"
	"strings"
	"time"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// the method is to convert to json format from one map sting got with -jsonpath
//
//nolint:unused
func convertLMtoJSON(content string) string {
	var jb strings.Builder
	jb.WriteString("[")
	items := strings.Split(strings.TrimSuffix(strings.TrimPrefix(content, "["), "]"), "map")
	for _, item := range items {
		if strings.Compare(item, "") == 0 {
			continue
		}
		kvs := strings.Fields(strings.TrimSuffix(strings.TrimPrefix(item, "["), "]"))
		jb.WriteString("{")
		for ki, kv := range kvs {
			p := strings.Split(kv, ":")
			jb.WriteString("\"" + p[0] + "\":")
			jb.WriteString("\"" + p[1] + "\"")
			if ki < len(kvs)-1 {
				jb.WriteString(", ")
			}
		}
		jb.WriteString("},")
	}
	return strings.TrimSuffix(jb.String(), ",") + "]"
}

// the method is to update z version of kube version of platform.
func GenerateUpdatedKubernatesVersion(oc *exutil.CLI) string {
	subKubeVersions := strings.Split(getKubernetesVersion(oc), ".")
	zVersion, _ := strconv.Atoi(subKubeVersions[1])
	subKubeVersions[1] = strconv.Itoa(zVersion + 1)
	return strings.Join(subKubeVersions[0:2], ".") + ".0"
}

// the method is to get kube versoin of the platform.
func getKubernetesVersion(oc *exutil.CLI) string {
	output, err := exutil.OcAction(oc, "version", exutil.AsAdmin, exutil.WithoutNamespace, "-o=json")
	o.Expect(err).NotTo(o.HaveOccurred())

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	o.Expect(err).NotTo(o.HaveOccurred())

	gitVersion := result["serverVersion"].(map[string]interface{})["gitVersion"]
	e2e.Logf("gitVersion is %v", gitVersion)
	return strings.TrimPrefix(gitVersion.(string), "v")
}

// the method is to create one resource with template
func ApplyResourceFromTemplate(oc *exutil.CLI, parameters ...string) error {
	var configFile string
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		output, err := oc.AsAdmin().Run("process").Args(parameters...).OutputToFile(exutil.GetRandomString() + "olm-config.json")
		if err != nil {
			e2e.Logf("the err:%v, and try next round", err)
			return false, nil
		}
		configFile = output
		return true, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("can not process %v", parameters))

	e2e.Logf("the file of resource is %s", configFile)
	return oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", configFile).Execute()
}

// the method is to check the presence of the resource
// AsAdmin means if taking admin to check it
// WithoutNamespace means if take WithoutNamespace() to check it.
// present means if you expect the resource presence or not. if it is ok, expect presence. if it is nok, expect not present.
func IsPresentResource(oc *exutil.CLI, AsAdmin bool, WithoutNamespace bool, present bool, parameters ...string) bool {

	return checkPresent(oc, 3, 70, AsAdmin, WithoutNamespace, present, parameters...)

}

// the method is basic method to check the presence of the resource
// AsAdmin means if taking admin to check it
// WithoutNamespace means if take WithoutNamespace() to check it.
// present means if you expect the resource presence or not. if it is ok, expect presence. if it is nok, expect not present.
func checkPresent(oc *exutil.CLI, intervalSec int, durationSec int, AsAdmin bool, WithoutNamespace bool, present bool, parameters ...string) bool {
	parameters = append(parameters, "--ignore-not-found")
	err := wait.PollUntilContextTimeout(context.TODO(), time.Duration(intervalSec)*time.Second, time.Duration(durationSec)*time.Second, false, func(ctx context.Context) (bool, error) {
		output, err := exutil.OcAction(oc, "get", AsAdmin, WithoutNamespace, parameters...)
		if err != nil {
			e2e.Logf("the get error is %v, and try next", err)
			return false, nil
		}
		if !present && strings.Compare(output, "") == 0 {
			return true, nil
		}
		if present && strings.Compare(output, "") != 0 {
			return true, nil
		}
		return false, nil
	})
	return err == nil
}

// the method is to patch one resource
// AsAdmin means if taking admin to patch it
// WithoutNamespace means if take WithoutNamespace() to patch it.
func PatchResource(oc *exutil.CLI, AsAdmin bool, WithoutNamespace bool, parameters ...string) {
	_, err := exutil.OcAction(oc, "patch", AsAdmin, WithoutNamespace, parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// the method is to execute something in pod to get output
// AsAdmin means if taking admin to execute it
// WithoutNamespace means if take WithoutNamespace() to execute it.
//
//nolint:unused
func execResource(oc *exutil.CLI, AsAdmin bool, WithoutNamespace bool, parameters ...string) string {
	var result string
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 6*time.Second, false, func(ctx context.Context) (bool, error) {
		output, err := exutil.OcAction(oc, "exec", AsAdmin, WithoutNamespace, parameters...)
		if err != nil {
			e2e.Logf("the exec error is %v, and try next", err)
			return false, nil
		}
		result = output
		return true, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("can not exec %v", parameters))
	e2e.Logf("the result of exec resource:%v", result)
	return result
}

// the method is to get something from resource. it is "oc get xxx" actaully
// AsAdmin means if taking admin to get it
// WithoutNamespace means if take WithoutNamespace() to get it.
func GetResource(oc *exutil.CLI, AsAdmin bool, WithoutNamespace bool, parameters ...string) string {
	var result string
	var err error
	err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 150*time.Second, false, func(ctx context.Context) (bool, error) {
		result, err = exutil.OcAction(oc, "get", AsAdmin, WithoutNamespace, parameters...)
		if err != nil {
			e2e.Logf("output is %v, error is %v, and try next", result, err)
			return false, nil
		}
		return true, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("can not get %v", parameters))
	e2e.Logf("$oc get %v, the returned resource:\n%v", parameters, result)
	return result
}

func GetResourceNoEmpty(oc *exutil.CLI, AsAdmin bool, WithoutNamespace bool, parameters ...string) string {
	var result string
	var err error
	err = wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 150*time.Second, false, func(ctx context.Context) (bool, error) {
		result, err = exutil.OcAction(oc, "get", AsAdmin, WithoutNamespace, parameters...)
		if err != nil || strings.TrimSpace(result) == "" {
			e2e.Logf("output is %v, error is %v, and try next", result, err)
			return false, nil
		}
		return true, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("can not get %v without empty", parameters))
	e2e.Logf("$oc get %v, the returned resource:\n%v", parameters, result)
	return result
}

// the method is to check one resource's attribution is expected or not.
// AsAdmin means if taking admin to check it
// WithoutNamespace means if take WithoutNamespace() to check it.
// isCompare means if Containing or exactly comparing. if it is Contain, it check result Contain content. if it is Compare, it Compare the result with content exactly.
// content is the substing to be expected
// the expect is ok, Contain or Compare result is OK for method == expect, no error raise. if not OK, error raise
// the expect is nok, Contain or Compare result is NOK for method == expect, no error raise. if OK, error raise
func expectedResource(oc *exutil.CLI, AsAdmin bool, WithoutNamespace bool, isCompare bool, content string, expect bool, parameters ...string) error {
	expectMap := map[bool]string{
		true:  "do",
		false: "do not",
	}

	cc := func(a, b string, ic bool) bool {
		bs := strings.Split(b, "+2+")
		ret := false
		for _, s := range bs {
			if (ic && strings.Compare(a, s) == 0) || (!ic && strings.Contains(a, s)) {
				ret = true
			}
		}
		return ret
	}
	e2e.Logf("Running: oc get AsAdmin(%t) WithoutNamespace(%t) %s", AsAdmin, WithoutNamespace, strings.Join(parameters, " "))

	// The detault timeout
	timeString := "300s"
	// extract the custom timeout
	if strings.Contains(content, "-TIME-WAIT-") {
		timeString = strings.Split(content, "-TIME-WAIT-")[1]
		content = strings.Split(content, "-TIME-WAIT-")[0]
		e2e.Logf("! reset the timeout to %s", timeString)
	}
	timeout, err := time.ParseDuration(timeString)
	if err != nil {
		e2e.Failf("! Fail to parse the timeout value:%s, err:%v", content, err)
	}

	return wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		output, err := exutil.OcAction(oc, "get", AsAdmin, WithoutNamespace, parameters...)
		if err != nil {
			e2e.Logf("the get error is %v, and try next", err)
			return false, nil
		}
		e2e.Logf("---> we %v expect value: %s, in returned value: %s", expectMap[expect], content, output)
		if isCompare && expect && cc(output, content, isCompare) {
			e2e.Logf("the output %s matches one of the content %s, expected", output, content)
			return true, nil
		}
		if isCompare && !expect && !cc(output, content, isCompare) {
			e2e.Logf("the output %s does not matche the content %s, expected", output, content)
			return true, nil
		}
		if !isCompare && expect && cc(output, content, isCompare) {
			e2e.Logf("the output %s Contains one of the content %s, expected", output, content)
			return true, nil
		}
		if !isCompare && !expect && !cc(output, content, isCompare) {
			e2e.Logf("the output %s does not Contain the content %s, expected", output, content)
			return true, nil
		}
		e2e.Logf("---> Not as expected! Return false")
		return false, nil
	})
}

// the method is to remove resource
// AsAdmin means if taking admin to remove it
// WithoutNamespace means if take WithoutNamespace() to remove it.
func removeResource(oc *exutil.CLI, AsAdmin bool, WithoutNamespace bool, parameters ...string) {
	output, err := exutil.OcAction(oc, "delete", AsAdmin, WithoutNamespace, parameters...)
	if err != nil && (strings.Contains(output, "NotFound") || strings.Contains(output, "No resources found")) {
		e2e.Logf("the resource is deleted already")
		return
	}
	o.Expect(err).NotTo(o.HaveOccurred())

	err = wait.PollUntilContextTimeout(context.TODO(), 4*time.Second, 160*time.Second, false, func(ctx context.Context) (bool, error) {
		output, err := exutil.OcAction(oc, "get", AsAdmin, WithoutNamespace, parameters...)
		if err != nil && (strings.Contains(output, "NotFound") || strings.Contains(output, "No resources found")) {
			e2e.Logf("the resource is delete successfully")
			return true, nil
		}
		return false, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("can not remove %v", parameters))
}

func ClusterPackageExists(oc *exutil.CLI, sub SubscriptionDescription) (bool, error) {
	msg, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", sub.OperatorPackage, "-n", sub.CatalogSourceNamespace).Output()
	if err != nil || strings.Contains(msg, "not found") {
		return false, err
	}
	return true, err
}

func ClusterPackageExistsInNamespace(oc *exutil.CLI, sub SubscriptionDescription, namespace string) (bool, error) {
	found := false
	var v []string
	var msg string
	var err error
	if namespace == "all" {
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "--all-namespaces", "-o=jsonpath={range .items[*]}{@.metadata.name}{\",\"}{@.metadata.labels.catalog}{\"\\n\"}{end}").Output()
	} else {
		msg, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", "-n", namespace, "-o=jsonpath={range .items[*]}{@.metadata.name}{\",\"}{@.metadata.labels.catalog}{\"\\n\"}{end}").Output()
	}
	if err == nil {
		for _, s := range strings.Fields(msg) {
			v = strings.Split(s, ",")
			if v[0] == sub.OperatorPackage && v[1] == sub.CatalogSourceName {
				found = true
				e2e.Logf("%v matches: %v", s, sub.OperatorPackage)
				break
			}
		}
	}
	if !found {
		e2e.Logf("%v was not found in \n%v", sub.OperatorPackage, msg)
	}
	return found, err
}

func SkipIfPackagemanifestNotExist(oc *exutil.CLI, packageName string) {
	if oc == nil {
		e2e.Logf("CLI client is nil")
		g.Skip("CLI client is nil, cannot check packagemanifest")
	}

	if packageName == "" {
		e2e.Logf("Package name is empty")
		g.Skip("Package name is empty, cannot check packagemanifest")
	}

	var output string
	var err error
	output, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifest", packageName, "--ignore-not-found").Output()

	if err != nil || strings.TrimSpace(output) == "" {
		e2e.Logf("Packagemanifest '%s' not found, error: %v", packageName, err)
		g.Skip(fmt.Sprintf("Packagemanifest '%s' not found. This test requires the packagemanifest to be available.", packageName))
	}
	e2e.Logf("Packagemanifest '%s' exists", packageName)
}

// Return a github client
func GithubClient() (context.Context, *http.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	return ctx, tc
}

// GetDirPath return a string of dir path
func GetDirPath(filePathStr string, filePre string) string {
	if !strings.Contains(filePathStr, "/") || filePathStr == "/" {
		return ""
	}
	dir, file := filepath.Split(filePathStr)
	if strings.HasPrefix(file, filePre) {
		return filePathStr
	}
	return GetDirPath(filepath.Dir(dir), filePre)
}

// DeleteDir delete the dir
func DeleteDir(filePathStr string, filePre string) bool {
	filePathToDelete := GetDirPath(filePathStr, filePre)
	if filePathToDelete == "" || !strings.Contains(filePathToDelete, filePre) {
		e2e.Logf("there is no such dir %s", filePre)
		return false
	}
	e2e.Logf("remove dir %s", filePathToDelete)
	if err := os.RemoveAll(filePathToDelete); err != nil {
		e2e.Logf("Failed to remove directory %s: %v", filePathToDelete, err)
		return false
	}
	if _, err := os.Stat(filePathToDelete); err == nil {
		e2e.Logf("delele dir %s failed", filePathToDelete)
		return false
	}
	return true
}

// CheckUpgradeStatus check upgrade status
func CheckUpgradeStatus(oc *exutil.CLI, expectedStatus string) {
	e2e.Logf("Check the Upgradeable status of the OLM, expected: %s", expectedStatus)
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
		upgradeable, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("co", "operator-lifecycle-manager", "-o=jsonpath={.status.conditions[?(@.type==\"Upgradeable\")].status}").Output()
		if err != nil {
			e2e.Failf("Fail to get the Upgradeable status of the OLM: %v", err)
		}
		if upgradeable != expectedStatus {
			return false, nil
		}
		e2e.Logf("The Upgraableable status should be %s, and get %s", expectedStatus, upgradeable)
		return true, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("Upgradeable status of the OLM %s is not expected", expectedStatus))
}

func notInList(target string, strArray []string) bool {
	for _, element := range strArray {
		if target == element {
			return false
		}
	}
	return true
}

func LogDebugInfo(oc *exutil.CLI, ns string, resource ...string) {
	for _, resourceIndex := range resource {
		e2e.Logf("oc get %s:", resourceIndex)
		output, _ := oc.AsAdmin().WithoutNamespace().Run("get").Args(resourceIndex, "-n", ns).Output()
		if strings.Contains(resourceIndex, "event") {
			var warningEventList []string
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "Warning") {
					warningStr := strings.Split(line, "Warning")[1]
					if notInList(warningStr, warningEventList) {
						warningEventList = append(warningEventList, "Warning"+warningStr)
					}
				}
			}
			e2e.Logf("Warning events: %s", strings.Join(warningEventList, "\n"))
		} else {
			e2e.Logf("Debug output: %s", output)
		}
	}
}
func IsSNOCluster(oc *exutil.CLI) bool {
	//Only 1 master, 1 worker node and with the same hostname.
	masterNodes, _ := exutil.GetClusterNodesBy(oc, "master")
	workerNodes, _ := exutil.GetClusterNodesBy(oc, "worker")
	e2e.Logf("masterNodes:%s, workerNodes:%s", masterNodes, workerNodes)
	if len(masterNodes) == 1 && len(workerNodes) == 1 && masterNodes[0] == workerNodes[0] {
		e2e.Logf("This is a SNO cluster")
		return true
	}
	return false
}

func AssertOrCheckMCP(oc *exutil.CLI, mcp string, is int, dm int, skip bool) {
	var machineCount string
	err := wait.PollUntilContextTimeout(context.TODO(), time.Duration(is)*time.Second, time.Duration(dm)*time.Minute, false, func(ctx context.Context) (bool, error) {
		machineCount, _ = oc.AsAdmin().WithoutNamespace().Run("get").Args("mcp", mcp, "-o=jsonpath={.status.machineCount}{\" \"}{.status.readyMachineCount}{\" \"}{.status.unavailableMachineCount}{\" \"}{.status.degradedMachineCount}").Output()
		indexCount := strings.Fields(machineCount)
		if strings.Compare(indexCount[0], indexCount[1]) == 0 && strings.Compare(indexCount[2], "0") == 0 && strings.Compare(indexCount[3], "0") == 0 {
			return true, nil
		}
		return false, nil
	})
	e2e.Logf("MachineCount:ReadyMachineCountunavailableMachineCountdegradedMachineCount: %v", machineCount)
	if err != nil {
		if skip {
			g.Skip(fmt.Sprintf("the mcp %v is not correct status, so skip it", machineCount))
		}
		exutil.AssertWaitPollNoErr(err, fmt.Sprintf("macineconfigpool %v update failed", mcp))
	}
}

func GetAllCSV(oc *exutil.CLI) []string {
	allCSVs, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "--all-namespaces", `-o=jsonpath={range .items[*]}{@.metadata.name}{","}{@.metadata.namespace}{","}{@.status.reason}{":"}{end}`).Output()
	if err != nil {
		e2e.Failf("!!! Couldn't get all CSVs:%v\n", err)
	}
	var csvListOutput []string
	csvList := strings.Split(allCSVs, ":")
	for _, csv := range csvList {
		if strings.Compare(csv, "") == 0 {
			continue
		}
		name := strings.Split(csv, ",")[0]
		ns := strings.Split(csv, ",")[1]
		reason := strings.Split(csv, ",")[2]
		if strings.Compare(reason, "Copied") == 0 {
			continue
		}
		csvListOutput = append(csvListOutput, ns+":"+name)
	}
	return csvListOutput
}

// ToDo:
func CreateCatalog(oc *exutil.CLI, catalogName, indexImage, catalogTemplate string) {
	catalog, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalog", catalogName).Output()
	if err != nil {
		if strings.Contains(catalog, "not found") {
			err = ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", catalogTemplate, "-p", fmt.Sprintf("NAME=%s", catalogName), fmt.Sprintf("IMAGE=%s", indexImage))
			if err != nil {
				e2e.Logf("Failed to create catalog %s: %s", catalogName, err)
				// we do not asser it here because it is possible race condition. it means two cases create it at same
				// time, and the second will raise error
			}
			// here we will assert if the catalog is created successfully with checking unpack status.
			// need to check unpack status before continue to use it
			err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
				phase, errPhase := oc.AsAdmin().WithoutNamespace().Run("get").Args("catalog", catalogName, "-o=jsonpath={.status.phase}").Output()
				if errPhase != nil {
					e2e.Logf("%v, next try", errPhase)
					return false, nil
				}
				if strings.Compare(phase, "Unpacked") == 0 {
					return true, nil
				}
				return false, nil
			})
			exutil.AssertWaitPollNoErr(err, "catalog unpack fails")

		} else {
			o.Expect(err).NotTo(o.HaveOccurred())
		}
	}
}

func GetCertRotation(oc *exutil.CLI, secretName, namespace string) (certsLastUpdated, certsRotateAt time.Time) {
	var certsEncoding string
	var err error
	err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 180*time.Second, false, func(ctx context.Context) (bool, error) {
		certsEncoding, err = oc.AsAdmin().WithoutNamespace().Run("get").Args("secret", secretName, "-n", namespace, "-o=jsonpath={.data}").Output()
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("Fail to get the certsEncoding, certsEncoding:%v, error:%v", certsEncoding, err))

	certs, err := base64.StdEncoding.DecodeString(gjson.Get(certsEncoding, `tls\.crt`).String())
	if err != nil {
		e2e.Failf("Fail to get the certs:%v, error:%v", certs, err)
	}
	block, _ := pem.Decode(certs)
	if block == nil {
		e2e.Failf("failed to parse certificate PEM")
	}
	dates, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		e2e.Failf("Fail to parse certificate:\n%v, error:%v", string(certs), err)
	}

	notBefore := dates.NotBefore
	notAfter := dates.NotAfter
	// code: https://github.com/jianzhangbjz/operator-framework-olm/commit/7275a55186a59fcb9845cbe3a9a99c56a7afbd1d
	duration, _ := time.ParseDuration("5m")
	secondsDifference := notBefore.Add(duration).Sub(notAfter).Seconds()
	if secondsDifference > 3 || secondsDifference < -3 {
		e2e.Failf("the duration is incorrect, notBefore:%v, notAfter:%v, secondsDifference:%v", notBefore, notAfter, secondsDifference)
	}

	g.By("rotation will be 1 minutes earlier than expiration")
	certsLastUpdadString, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.certsLastUpdated}").Output()
	if err != nil {
		e2e.Failf("Fail to get certsLastUpdated, error:%v", err)
	}
	certsRotateAtString, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("csv", "packageserver", "-n", "openshift-operator-lifecycle-manager", "-o=jsonpath={.status.certsRotateAt}").Output()
	if err != nil {
		e2e.Failf("Fail to get certsRotateAt, error:%v", err)
	}
	duration2, _ := time.ParseDuration("4m")
	certsLastUpdated, _ = time.Parse(time.RFC3339, certsLastUpdadString)
	certsRotateAt, _ = time.Parse(time.RFC3339, certsRotateAtString)
	// certsLastUpdated:2022-08-23 08:59:45
	// certsRotateAt:2022-08-23 09:03:44
	// due to https://issues.redhat.com/browse/OCPBUGS-444, there is a 1s difference, so here check if seconds difference in 3s.
	secondsDifference = certsLastUpdated.Add(duration2).Sub(certsRotateAt).Seconds()
	if secondsDifference > 3 || secondsDifference < -3 {
		e2e.Failf("the certsRotateAt beyond 3s than expected, certsLastUpdated:%v, certsRotateAt:%v", certsLastUpdated, certsRotateAt)
	}
	return certsLastUpdated, certsRotateAt
}

// Common user use oc client apply yaml template
func ApplyResourceFromTemplateOnMicroshift(oc *exutil.CLI, parameters ...string) error {
	configFile := exutil.ParameterizedTemplateByReplaceToFile(oc, parameters...)
	e2e.Logf("the file of resource is %s", configFile)
	return oc.WithoutNamespace().Run("apply").Args("-f", configFile).Execute()
}

func GetCatsrc(oc *exutil.CLI, catsrc string, operator string) string {
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("packagemanifests.packages.operators.coreos.com",
		"--selector=catalog="+catsrc, "--field-selector", "metadata.name="+operator,
		"-o=jsonpath={..status.catalogSource}").Output()
	if err != nil {
		e2e.Logf("can not get catsrc for %s with %v", operator, err)
		return ""
	}
	return output
}

// return a map that pod's image is key and imagePullPolicy is value
func GetPodImageAndPolicy(oc *exutil.CLI, podName, project string) (imageMap map[string]string) {
	imageMap = make(map[string]string)
	if podName == "" || project == "" {
		return imageMap
	}
	Containers := []string{"initContainers", "Containers"}
	for _, v := range Containers {
		imageNameSlice := []string{}
		imagePullPolicySlice := []string{}
		jsonPathImage := fmt.Sprintf("-o=jsonpath={.spec.%s[*].image}", v)
		jsonPathPolicy := fmt.Sprintf("-o=jsonpath={.spec.%s[*].imagePullPolicy}", v)

		imageNames, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(podName, jsonPathImage, "-n", project).Output()
		// sometimes some job's pod maybe deleted so skip it
		if err != nil {
			if !strings.Contains(imageNames, "NotFound") {
				e2e.Failf("Fail to get image(%s), error:%s", podName, imageNames)
			}
		} else {
			imageNameSlice = strings.Split(imageNames, " ")
		}

		imagePullPolicys, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(podName, jsonPathPolicy, "-n", project).Output()
		if err != nil {
			if !strings.Contains(imagePullPolicys, "NotFound") {
				e2e.Failf("Fail to get imagePullPolicy(%s), error:%s", podName, imagePullPolicys)
			}
		} else {
			imagePullPolicySlice = strings.Split(imagePullPolicys, " ")
		}

		if len(imageNameSlice) < 1 || len(imagePullPolicySlice) < 1 {
			continue
		}
		for i := 0; i < len(imageNameSlice); i++ {
			if _, ok := imageMap[imageNameSlice[i]]; !ok {
				imageMap[imageNameSlice[i]] = imagePullPolicySlice[i]
			}
		}
	}
	return imageMap
}

// return a pod slice
func GetProjectPods(oc *exutil.CLI, project string) (podSlice []string) {
	podSlice = []string{}
	if project == "" {
		return podSlice
	}
	pods, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("pods", "-o", "name", "-n", project).Output()
	if err != nil {
		e2e.Failf("Fail to get %s pods, error:%v", project, err)
	}
	podSlice = strings.Split(pods, "\n")
	return podSlice
}

func ClusterHasEnabledFIPS(oc *exutil.CLI) bool {
	firstNode, err := exutil.GetFirstMasterNode(oc)
	msgIfErr := fmt.Sprintf("ERROR Could not get first node to check FIPS '%v' %v", firstNode, err)
	o.Expect(err).NotTo(o.HaveOccurred(), msgIfErr)
	o.Expect(firstNode).NotTo(o.BeEmpty(), msgIfErr)
	// hardcode the default project since its enforce is privileged as default
	fipsModeStatus, err := oc.AsAdmin().Run("debug").Args("-n", "default", "node/"+firstNode, "--", "chroot", "/host", "fips-mode-setup", "--check").Output()
	msgIfErr = fmt.Sprintf("ERROR Could not check FIPS on node %v: '%v' %v", firstNode, fipsModeStatus, err)
	o.Expect(err).NotTo(o.HaveOccurred(), msgIfErr)
	o.Expect(fipsModeStatus).NotTo(o.BeEmpty(), msgIfErr)

	// This will be true or false
	return strings.Contains(fipsModeStatus, "FIPS mode is enabled.")
}

func BeforeTargetTime(logText, targetText string) bool {
	e2e.Logf("Get the x509 log text: %s", logText)
	var logDateTime time.Time
	var err error
	var found bool

	// try RFC3339
	if match := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z`).FindString(logText); match != "" {
		logDateTime, err = time.Parse(time.RFC3339Nano, match)
		if err != nil {
			e2e.Failf("fail to parse reRFC3339 time: %v", err)
		}
		found = true
	}

	// try match（2025/04/12 19:20:12）
	if !found {
		if match := regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`).FindString(logText); match != "" {
			logDateTime, err = time.Parse("2006/01/02 15:04:05", match)
			if err != nil {
				e2e.Failf("fail to parse standard time: %v", err)
			}
			found = true
		}
	}

	// try Kubernetes（E0321 19:45:14.489789）
	if !found {
		if matches := regexp.MustCompile(`E(\d{4}) (\d{2}:\d{2}:\d{2}\.\d+)`).FindStringSubmatch(logText); len(matches) == 3 {
			monthDay := matches[1]
			timeStr := matches[2]

			month, _ := strconv.Atoi(monthDay[:2])
			day, _ := strconv.Atoi(monthDay[2:])
			logTime, err := time.Parse("15:04:05.999999", timeStr)
			if err != nil {
				e2e.Failf("fail to parse reK8s time:: %v", err)
			}
			logDateTime = time.Date(time.Now().Year(), time.Month(month), day,
				logTime.Hour(), logTime.Minute(), logTime.Second(), logTime.Nanosecond(), time.UTC)
			found = true
		}
	}

	// all not found
	if !found {
		e2e.Failf("fail to parse any time format: %s", logText)
	}

	targetTime, err := time.Parse(time.RFC3339, targetText)
	if err != nil {
		e2e.Failf("Fail to parse target time(RFC3339): %v", err)
	}

	e2e.Logf("log time: %s", logDateTime)
	e2e.Logf("target time: %s", targetTime)

	if logDateTime.Before(targetTime) {
		e2e.Logf("The log time is before the target time")
		return true
	} else {
		e2e.Logf("The log time is after the target time")
		return false
	}
}

type Metrics struct {
	csvCount              int
	csvUpgradeCount       int
	catalogSourceCount    int
	installPlanCount      int
	subscriptionCount     int
	subscriptionSyncTotal int
}

// PrometheusQueryResult the prometheus query result
type PrometheusQueryResult struct {
	Data struct {
		Result []struct {
			Metric struct {
				Name      string `json:"__name__"`
				Approval  string `json:"approval"`
				Channel   string `json:"channel"`
				Container string `json:"Container"`
				Endpoint  string `json:"endpoint"`
				Installed string `json:"installed"`
				Instance  string `json:"instance"`
				Job       string `json:"job"`
				SrcName   string `json:"name"`
				Namespace string `json:"namespace"`
				Package   string `json:"package"`
				Pod       string `json:"pod"`
				Service   string `json:"service"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
		ResultType string `json:"resultType"`
	} `json:"data"`
	Status string `json:"status"`
}

func GetMetrics(oc *exutil.CLI, olmToken string, data PrometheusQueryResult, metrics Metrics, subName, prometheusPodIP string) Metrics {
	// don't show the token info even if the token is transitory
	oc.NotShowInfo()
	defer oc.SetShowInfo()

	args := []string{"-n", "openshift-monitoring", "prometheus-k8s-0", "-i", "--", "curl"}
	extraEnvUnset := []string{"env", "-u", "http_proxy", "-u", "https_proxy", "-u", "HTTP_PROXY", "-u", "HTTPS_PROXY", "--noproxy", "'*'"}
	if IsIPv6(prometheusPodIP) {
		prometheusPodIP = fmt.Sprintf("[%s]", prometheusPodIP)
		args = append(args, extraEnvUnset...)
	}
	e2e.Logf("openshift-monitoring/prometheus-k8s-0 pod IP:%s", prometheusPodIP)
	metricsCon := []string{"csv_count", "csv_upgrade_count", "catalog_source_count", "install_plan_count", "subscription_count", "subscription_sync_total"}
	for _, metric := range metricsCon {
		waitErr := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
			queryContent := fmt.Sprintf("https://%s:9091/api/v1/query?query=%s", prometheusPodIP, metric)
			execArgs := append([]string{}, args...)
			execArgs = append(execArgs, "-k", "-H", fmt.Sprintf("Authorization: Bearer %v", olmToken), queryContent)
			msg, _, err := oc.AsAdmin().WithoutNamespace().Run("exec").Args(execArgs...).Outputs()
			e2e.Logf("%s, err:%v, msg:%v", metric, err, msg)
			if msg == "" {
				return false, nil
			}
			if err := json.Unmarshal([]byte(msg), &data); err != nil {
				e2e.Logf("Failed to unmarshal JSON response: %v", err)
				return false, nil
			}
			//subscription_sync_total, err:<nil>, msg:{"status":"success","data":{"resultType":"vector","result":[]}}
			if len(data.Data.Result) < 1 || len(data.Data.Result[0].Value) < 2 {
				if metric == "subscription_sync_total" {
					return true, nil
				}
				return false, nil
			}
			if metric == "subscription_sync_total" {
				metrics.subscriptionSyncTotal = 0
				for i := range data.Data.Result {
					if strings.Contains(data.Data.Result[i].Metric.SrcName, subName) {
						metrics.subscriptionSyncTotal, _ = strconv.Atoi(data.Data.Result[i].Value[1].(string))
					}
				}
			} else {
				switch metric {
				case "csv_count":
					metrics.csvCount, _ = strconv.Atoi(data.Data.Result[0].Value[1].(string))
				case "csv_upgrade_count":
					metrics.csvUpgradeCount, _ = strconv.Atoi(data.Data.Result[0].Value[1].(string))
				case "catalog_source_count":
					metrics.catalogSourceCount, _ = strconv.Atoi(data.Data.Result[0].Value[1].(string))
				case "install_plan_count":
					metrics.installPlanCount, _ = strconv.Atoi(data.Data.Result[0].Value[1].(string))
				case "subscription_count":
					metrics.subscriptionCount, _ = strconv.Atoi(data.Data.Result[0].Value[1].(string))
				}
			}
			return true, nil
		})
		exutil.AssertWaitPollNoErr(waitErr, fmt.Sprintf("failed to query %s", metric))
	}
	return metrics
}

// HasExternalNetworkAccess tests network connectivity from a cluster master node
// by attempting to access an external container registry (quay.io).
// This method uses DebugNodeWithChroot to avoid creating pods and pulling images,
// which would fail in disconnected environments.
//
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if external network access is available, false otherwise
func HasExternalNetworkAccess(oc *exutil.CLI) bool {
	if oc == nil {
		e2e.Logf("CLI client is nil, assuming connected environment")
		return true
	}

	e2e.Logf("Testing external network connectivity from master node using DebugNodeWithChroot")

	masterNode, masterErr := exutil.GetFirstMasterNode(oc)
	if masterErr != nil {
		e2e.Logf("Failed to get master node: %v", masterErr)
		g.Skip(fmt.Sprintf("Cannot determine network connectivity: %v", masterErr))
	}

	// Test connectivity to quay.io (container registry)
	// Use timeout to avoid hanging, and redirect output to check connection status
	// Note: In disconnected environments, curl will fail and bash will return non-zero exit code,
	// causing DebugNodeWithChroot to return an error. We ignore this error and rely on output checking.
	cmd := `timeout 10 curl -k https://quay.io > /dev/null 2>&1; [ $? -eq 0 ] && echo "connected"`
	output, _ := exutil.DebugNodeWithOptionsAndChroot(oc, masterNode, []string{"--to-namespace=default"}, "bash", "-c", cmd)

	// Check if the output contains "connected"
	// - Connected environment: curl succeeds -> echo "connected" -> output contains "connected"
	// - Disconnected environment: curl fails -> no echo -> output empty or only debug messages
	if strings.Contains(output, "connected") {
		e2e.Logf("External network connectivity test succeeded (output: %s), cluster can access quay.io", strings.TrimSpace(output))
		return true
	}

	e2e.Logf("External network connectivity test failed (output: %s), cluster cannot access quay.io", strings.TrimSpace(output))
	return false
}

// IsProxyCluster checks whether the cluster is configured with HTTP/HTTPS proxy.
// Proxy clusters are treated as connected environments since they can access external networks through the proxy.
//
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - true if cluster has HTTP or HTTPS proxy configured in status
//   - false if no proxy is configured
//
// Behavior:
//   - Skips the test if oc is nil or if error occurs while checking proxy configuration
func IsProxyCluster(oc *exutil.CLI) bool {
	if oc == nil {
		e2e.Logf("CLI client is nil, cannot check proxy configuration")
		g.Skip("CLI client is nil, cannot check proxy configuration")
	}

	// Get proxy status in one call to check both httpProxy and httpsProxy
	// Format: {"httpProxy":"<value>","httpsProxy":"<value>"}
	proxyStatus, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("proxy", "cluster", "-o=jsonpath={.status}").Output()
	if err != nil {
		e2e.Logf("Failed to get proxy status: %v", err)
		g.Skip(fmt.Sprintf("cannot get proxy status: %v", err))
	}

	// If either httpProxy or httpsProxy is configured, the status will contain http
	// Connected cluster status is empty "{}"
	// Proxy cluster status contains "httpProxy" or "httpsProxy" fields with non-empty values
	if strings.Contains(proxyStatus, "httpProxy") || strings.Contains(proxyStatus, "httpsProxy") {
		e2e.Logf("Proxy cluster detected")
		return true
	}

	e2e.Logf("No proxy configuration detected in cluster (status=%s)", proxyStatus)
	return false
}

// ValidateAccessEnvironment checks if the cluster is in a disconnected environment
// and validates that required mirror configurations (ImageTagMirrorSet) are present.
// This should be called at the beginning of test cases that support disconnected environments.
//
// The function recognizes three types of cluster network access:
//  1. Connected: Direct access to external networks (no proxy, no disconnected)
//  2. Proxy: Access through HTTP/HTTPS proxy (treated as connected)
//  3. Disconnected: No external access, requires ImageTagMirrorSet for image mirroring
//
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Behavior:
//   - Skips the test if master node cannot be accessed (cannot determine environment)
//   - Returns immediately if proxy cluster detected (no mirror validation needed)
//   - Skips the test if in disconnected environment but ImageTagMirrorSet is not configured
//   - Continues normally if in connected environment or disconnected with proper configuration
//
// Usage:
//
//	g.It("test case supporting disconnected", func() {
//	    olmv0util.ValidateAccessEnvironment(oc)
//	    // rest of test code
//	})
func ValidateAccessEnvironment(oc *exutil.CLI) {
	// First check if this is a proxy cluster
	// Proxy clusters can access external networks through proxy, so they don't need mirror validation
	if IsProxyCluster(oc) {
		e2e.Logf("Proxy cluster detected, treating as connected environment (no mirror validation needed)")
		return
	}

	// Check if we can access external network directly
	hasNetwork := HasExternalNetworkAccess(oc)

	// If connected (and not proxy, already checked above), no validation needed
	if hasNetwork {
		e2e.Logf("Cluster has external network access (connected environment), no mirror validation needed")
		return
	}

	// In disconnected environment (not proxy, no external access), check for required ImageTagMirrorSet
	e2e.Logf("Cluster is in disconnected environment, validating ImageTagMirrorSet configuration")

	// Check if ImageTagMirrorSet "image-policy-aosqe" exists
	itmsOutput, itmsErr := oc.AsAdmin().WithoutNamespace().Run("get").Args("imagetagmirrorset", "image-policy-aosqe", "--ignore-not-found").Output()
	if itmsErr != nil || !strings.Contains(itmsOutput, "image-policy-aosqe") {
		g.Skip(fmt.Sprintf("Disconnected environment detected but ImageTagMirrorSet 'image-policy-aosqe' is not configured. "+
			"This test requires proper mirror configuration to run in disconnected clusters. "+
			"ITMS check result: output=%q, error=%v", itmsOutput, itmsErr))
	}

	e2e.Logf("Disconnected environment validation passed: ImageTagMirrorSet 'image-policy-aosqe' is configured")
}

// IsIPv6 check if the string is an IPv6 address.
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

// RemoveNamespace is the method to delete ns with namespace parameter if it exists
func RemoveNamespace(namespace string, oc *exutil.CLI) {
	_, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("ns", namespace).Output()

	if err == nil {
		_, err := oc.WithoutNamespace().AsAdmin().Run("delete").Args("ns", namespace).Output()
		o.Expect(err).NotTo(o.HaveOccurred())
	}
}
