package olm

import "context"

// OperatorPlugin provides a simple interface
// that can be used to extend the olm operator's functionality
type OperatorPlugin interface {
	Init(ctx context.Context, config *operatorConfig, operator *Operator) error
}
