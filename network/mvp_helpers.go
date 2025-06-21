package network

import "fmt"

// InputNeuronIDs_MVP_Preview returns a small preview of input neuron IDs.
func (cn *CrowNet) InputNeuronIDs_MVP_Preview(count int) string {
	if len(cn.InputNeuronIDs) == 0 {
		return "[]"
	}
	max := count
	if len(cn.InputNeuronIDs) < count {
		max = len(cn.InputNeuronIDs)
	}
	return fmt.Sprintf("%v", cn.InputNeuronIDs[:max])
}

// OutputNeuronIDs_MVP_Preview returns a small preview of output neuron IDs.
// For the digit task, it's important to show all 10 if they exist.
func (cn *CrowNet) OutputNeuronIDs_MVP_Preview(count int) string {
	if len(cn.OutputNeuronIDs) == 0 {
		return "[]"
	}
	// If count is 10 (for digits), try to show all 10.
	max := count
	if count == 10 && len(cn.OutputNeuronIDs) < 10 {
		max = len(cn.OutputNeuronIDs) // show what's available if less than 10
	} else if len(cn.OutputNeuronIDs) < count {
		max = len(cn.OutputNeuronIDs)
	}

	if max == 0 && len(cn.OutputNeuronIDs) > 0 { // ensure at least one is shown if list not empty
		max = 1
	}

	return fmt.Sprintf("%v", cn.OutputNeuronIDs[:max])
}
