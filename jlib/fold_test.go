package jlib

import (
	"github.com/goccy/go-json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFold(t *testing.T) {
	t.Run("fold a []interface", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{"Amount": 8},
			map[string]interface{}{"Amount": -1},
			map[string]interface{}{"Amount": 0},
			map[string]interface{}{"Amount": -3},
			map[string]interface{}{"Amount": -4},
		}

		output, err := FoldArray(input)
		assert.NoError(t, err)

		outputJSON, err := json.Marshal(output)
		assert.NoError(t, err)

		assert.Equal(t, string(outputJSON), `[[{"Amount":8}],[{"Amount":8},{"Amount":-1}],[{"Amount":8},{"Amount":-1},{"Amount":0}],[{"Amount":8},{"Amount":-1},{"Amount":0},{"Amount":-3}],[{"Amount":8},{"Amount":-1},{"Amount":0},{"Amount":-3},{"Amount":-4}]]`)
	})

	t.Run("fold - not an []interface{}", func(t *testing.T) {
		input := "testing"

		output, err := FoldArray(input)
		assert.Error(t, err)

		outputJSON, err := json.Marshal(output)
		assert.NoError(t, err)

		assert.Equal(t, string(outputJSON), `null`)
	})
}
