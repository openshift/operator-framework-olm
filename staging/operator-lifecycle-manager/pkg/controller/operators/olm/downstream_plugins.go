package olm

func init() {
	operatorPlugIns = []OperatorPlugin{
		// csv namespace labeler
		&CSVNamespaceLabelerPlugin{},
	}
}
