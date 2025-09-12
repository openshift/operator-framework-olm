package specs

import (
	"context"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("[sig-operator][Jira:OLM] OLMv0", func() {
	g.It("should pass a trivial sanity check", func(ctx context.Context) {
		o.Expect(len("test")).To(o.BeNumerically(">", 0))
	})
})
