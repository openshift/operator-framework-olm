package olmv0util

import (
	"fmt"
	"strings"

	o "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

type Port struct {
	Port     interface{} // int or string
	Protocol string
}
type Selector struct {
	NamespaceLabels map[string]string
	PodLabels       map[string]string
}
type IngressRule struct {
	Ports     []Port
	Selectors []Selector // maps to .from entries
}
type EgressRule struct {
	Ports     []Port
	Selectors []Selector // maps to .to entries
}
type NpExpecter struct {
	Name              string
	Namespace         string
	ExpectIngress     []IngressRule
	ExpectEgress      []EgressRule
	ExpectSelector    map[string]string
	ExpectPolicyTypes []string
}

// VerifySelector validates that the NetworkPolicy's podSelector matches the expected selector labels
// Parameters:
//   - specs: JSON string Containing the NetworkPolicy specification
//   - sel: map of expected selector labels (key-value pairs)
//   - name: NetworkPolicy name for error reporting
func VerifySelector(specs string, sel map[string]string, name string) {
	o.Expect(strings.TrimSpace(specs)).NotTo(o.BeEmpty(), "NetworkPolicy specification cannot be empty")
	o.Expect(strings.TrimSpace(name)).NotTo(o.BeEmpty(), "NetworkPolicy name cannot be empty")
	m := gjson.Get(specs, "podSelector.matchLabels").Map()
	o.Expect(len(m)).To(o.Equal(len(sel)), "%s: podSelector keys mismatch", name)
	for k, v := range sel {
		o.Expect(m[k].String()).To(o.Equal(v), "%s: podSelector[%s] mismatch", name, k)
	}
}

// VerifyPolicyTypes validates that the NetworkPolicy's policyTypes match the expected types
// Parameters:
//   - specs: JSON string Containing the NetworkPolicy specification
//   - types: slice of expected policy types (e.g., ["Ingress", "Egress"])
//   - name: NetworkPolicy name for error reporting
func VerifyPolicyTypes(specs string, types []string, name string) {
	o.Expect(strings.TrimSpace(specs)).NotTo(o.BeEmpty(), "NetworkPolicy specification cannot be empty")
	o.Expect(strings.TrimSpace(name)).NotTo(o.BeEmpty(), "NetworkPolicy name cannot be empty")
	o.Expect(types).NotTo(o.BeNil(), "policy types cannot be nil")
	arr := gjson.Get(specs, "policyTypes").Array()
	o.Expect(len(arr)).To(o.Equal(len(types)), "%s: policyTypes count mismatch", name)
	for i, exp := range types {
		o.Expect(arr[i].String()).To(o.Equal(exp), "%s: policyTypes[%d] mismatch", name, i)
	}
}

// VerifyIngress validates that the NetworkPolicy's ingress rules match the expected configuration
// Parameters:
//   - specs: JSON string Containing the NetworkPolicy specification
//   - rules: slice of expected IngressRule structures with ports and selectors
//   - name: NetworkPolicy name for error reporting
func VerifyIngress(specs string, rules []IngressRule, name string) {
	o.Expect(strings.TrimSpace(specs)).NotTo(o.BeEmpty(), "NetworkPolicy specification cannot be empty")
	o.Expect(strings.TrimSpace(name)).NotTo(o.BeEmpty(), "NetworkPolicy name cannot be empty")
	if len(rules) == 0 {
		o.Expect(gjson.Get(specs, "ingress").Exists()).
			To(o.BeFalse(), "%s: expected no ingress rules", name)
		return
	}
	o.Expect(int(gjson.Get(specs, "ingress.#").Int())).
		To(o.Equal(len(rules)), "%s: ingress rules count", name)

	for idx, rule := range rules {
		base := fmt.Sprintf("ingress.%d", idx)
		// ports

		if len(rule.Ports) == 1 && rule.Ports[0].Port == nil {
			o.Expect(gjson.Get(specs, base+".ports").Exists()).
				To(o.BeFalse(), "%s: rule %d empty ingress should have no ports", name, idx)
			o.Expect(gjson.Get(specs, base+".from").Exists()).
				To(o.BeFalse(), "%s: rule %d empty ingress should have no from", name, idx)
			continue
		}
		o.Expect(int(gjson.Get(specs, base+".ports.#").Int())).
			To(o.Equal(len(rule.Ports)), "%s: rule %d ingress ports count", name, idx)
		for i, exp := range rule.Ports {
			got := gjson.Get(specs, fmt.Sprintf("%s.ports.%d.port", base, i)).String()
			o.Expect(got).
				To(o.Equal(fmt.Sprint(exp.Port)), "%s: rule %d port #%d mismatch", name, idx, i)
			proto := gjson.Get(specs, fmt.Sprintf("%s.ports.%d.protocol", base, i)).String()
			o.Expect(proto).
				To(o.Equal(exp.Protocol), "%s: rule %d protocol #%d mismatch", name, idx, i)
		}
		// selectors
		if len(rule.Selectors) > 0 {
			for si, sel := range rule.Selectors {
				path := fmt.Sprintf("%s.from.%d", base, si)
				if sel.NamespaceLabels != nil {
					prefix := path + ".namespaceSelector.matchLabels"
					for k, v := range sel.NamespaceLabels {
						o.Expect(gjson.Get(specs, prefix+"."+escapeJSONPath(k)).String()).
							To(o.Equal(v), "%s: ingress rule %d selector %d nsLabel %s mismatch", name, idx, si, k)
					}
				}
				if sel.PodLabels != nil {
					prefix := path + ".podSelector.matchLabels"
					for k, v := range sel.PodLabels {
						o.Expect(gjson.Get(specs, prefix+"."+k).String()).
							To(o.Equal(v), "%s: ingress rule %d selector %d podLabel %s mismatch", name, idx, si, k)
					}
				}
			}
		}
	}
}

// VerifyEgress validates that the NetworkPolicy's egress rules match the expected configuration
// Parameters:
//   - specs: JSON string Containing the NetworkPolicy specification
//   - rules: slice of expected EgressRule structures with ports and selectors
//   - name: NetworkPolicy name for error reporting
func VerifyEgress(specs string, rules []EgressRule, name string) {
	o.Expect(strings.TrimSpace(specs)).NotTo(o.BeEmpty(), "NetworkPolicy specification cannot be empty")
	o.Expect(strings.TrimSpace(name)).NotTo(o.BeEmpty(), "NetworkPolicy name cannot be empty")
	if len(rules) == 0 {
		o.Expect(gjson.Get(specs, "egress").Exists()).
			To(o.BeFalse(), "%s: expected no egress rules", name)
		return
	}
	o.Expect(int(gjson.Get(specs, "egress.#").Int())).
		To(o.Equal(len(rules)), "%s: egress rules count", name)

	for idx, rule := range rules {
		base := fmt.Sprintf("egress.%d", idx)
		// ports
		if len(rule.Ports) == 1 && rule.Ports[0].Port == nil {
			// empty rule
			o.Expect(gjson.Get(specs, base+".ports").Exists()).
				To(o.BeFalse(), "%s: rule %d empty egress should have no ports", name, idx)
			o.Expect(gjson.Get(specs, base+".to").Exists()).
				To(o.BeFalse(), "%s: rule %d empty egress should have no to", name, idx)
			continue
		}
		o.Expect(int(gjson.Get(specs, base+".ports.#").Int())).
			To(o.Equal(len(rule.Ports)), "%s: rule %d egress ports count", name, idx)
		for i, exp := range rule.Ports {
			got := gjson.Get(specs, fmt.Sprintf("%s.ports.%d.port", base, i)).String()
			o.Expect(got).
				To(o.Equal(fmt.Sprint(exp.Port)), "%s: rule %d port #%d mismatch", name, idx, i)
			proto := gjson.Get(specs, fmt.Sprintf("%s.ports.%d.protocol", base, i)).String()
			o.Expect(proto).
				To(o.Equal(exp.Protocol), "%s: rule %d protocol #%d mismatch", name, idx, i)
		}
		// selectors
		if len(rule.Selectors) > 0 {
			for si, sel := range rule.Selectors {
				path := fmt.Sprintf("%s.to.%d", base, si)
				if sel.NamespaceLabels != nil {
					prefix := path + ".namespaceSelector.matchLabels"
					for k, v := range sel.NamespaceLabels {
						o.Expect(gjson.Get(specs, prefix+"."+escapeJSONPath(k)).String()).
							To(o.Equal(v), "%s: egress rule %d selector %d nsLabel %s mismatch", name, idx, si, k)
					}
				}
				if sel.PodLabels != nil {
					prefix := path + ".podSelector.matchLabels"
					for k, v := range sel.PodLabels {
						o.Expect(gjson.Get(specs, prefix+"."+k).String()).
							To(o.Equal(v), "%s: egress rule %d selector %d podLabel %s mismatch", name, idx, si, k)
					}
				}
			}
		}
	}
}

// escapeJSONPath escapes special characters in JSON path keys to prevent parsing errors
// Parameters:
//   - key: string key that may Contain special characters like '.' or '/'
//
// Returns:
//   - string: escaped key safe for use in JSON path expressions
func escapeJSONPath(key string) string {
	if key == "" {
		return ""
	}
	var esc strings.Builder
	for _, c := range key {
		if c == '.' || c == '/' {
			esc.WriteString(`\` + string(c))
		} else {
			esc.WriteRune(c)
		}
	}
	return esc.String()
}
