package trigger

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetKey(t *testing.T) {

	js := `
{
  "Inputs": [
    {
      "ParameterName": "account",
      "ParameterType": "address",
      "ParameterValue": "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984"
    }
  ],
  "Outputs": [
  ],
  "ContractABI": "",
  "ContractAdd": "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
  "TriggerName": "balance of",
  "TriggerType": "WatchContracts",
  "FunctionName": "balanceOf"
}
`
	tg, err := NewTriggerFromJson(js)
	assert.NoError(t, err)
	assert.Equal(t, "balanceOf+0x1f9840a85d5af5bf1d1762f925bdaddc4201f984+0x1f9840a85d5af5bf1d1762f925bdaddc4201f984", tg.getKey())
}
