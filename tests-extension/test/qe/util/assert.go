package util

import (
	"fmt"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// AssertWaitPollNoErr asserts that a wait.Poll operation completed without error
// Parameters:
//   - e: error returned from Wait.Poll operation
//   - msg: custom message to display when timeout occurs
//
// This function handles common timeout errors by replacing generic messages with more specific ones.
// It will cause the test to fail if any error is encountered.
func AssertWaitPollNoErr(e error, msg string) {
	if e == nil {
		return
	}

	if msg == "" {
		msg = "operation timed out"
	}

	var err error
	errorMsg := e.Error()

	// Use direct string comparison instead of strings.Compare for better performance
	if errorMsg == "timed out waiting for the condition" || errorMsg == "context deadline exceeded" {
		err = fmt.Errorf("case: %v\nerror: %s", g.CurrentSpecReport().FullText(), msg)
	} else {
		err = fmt.Errorf("case: %v\nerror: %s", g.CurrentSpecReport().FullText(), errorMsg)
	}
	o.Expect(err).NotTo(o.HaveOccurred())
}

// AssertWaitPollWithErr asserts that a wait.Poll operation failed with an expected error
// Parameters:
//   - e: error returned from Wait.Poll operation
//   - msg: message explaining why an error was expected
//
// This function expects an error to occur. If no error is returned, the test will fail.
// If an error is returned as expected, it logs the error and continues.
func AssertWaitPollWithErr(e error, msg string) {
	if e != nil {
		e2e.Logf("Expected error occurred: %v", e)
		return
	}

	if msg == "" {
		msg = "unknown reason"
	}

	err := fmt.Errorf("case: %v\nexpected error not received because of %v", g.CurrentSpecReport().FullText(), msg)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// OrFail processes function return values and fails the test if any error is encountered
// Parameters:
//   - vals: variadic arguments containing return values from a function call
//
// Returns:
//   - T: the first non-error value cast to type T
//
// Example usage:
//
//	func getValue() (string, error) { return "hello", nil }
//	value := OrFail[string](getValue())
//
// This function expects at least one argument and will fail the test if no values are provided.
// It will cause the test to fail if any of the values is a non-nil error.
func OrFail[T any](vals ...any) T {
	o.Expect(len(vals)).NotTo(o.BeZero(), "OrFail: no values provided")

	// Check for errors in any position
	for _, val := range vals {
		if err, ok := val.(error); ok && err != nil {
			o.ExpectWithOffset(1, err).NotTo(o.HaveOccurred())
		}
	}

	// Safely cast the first value to type T
	result, ok := vals[0].(T)
	o.Expect(ok).To(o.BeTrue(), "OrFail: first value cannot be cast to type %T", result)

	return result
}
