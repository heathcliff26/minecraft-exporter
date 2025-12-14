package save

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSaveCollector(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c, err := NewSaveCollector("not-a-path", "test-instance", false)
	assert.Error(err, "Should not create SaveCollector with invalid path")
	assert.Nil(c, "SaveCollector should be nil on error")

	c, err = NewSaveCollector("./testdata/1.20", "test-instance", false)
	assert.NoError(err, "Should create SaveCollector with valid path")
	require.NotNil(c, "SaveCollector should not be nil on success")
	assert.Equal("test-instance", c.Instance, "Instance label should be set correctly")
	assert.NotNil(c.save, "Should have a save instance")
	assert.NotNil(c.uuidCache, "Should have a uuidCache instance")
}

func TestCollectHasCommonLabels(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c, err := NewSaveCollector("./testdata/1.20", "test-instance", false)
	require.NoError(err)

	ch := make(chan prometheus.Metric)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	for metric := range ch {
		desc := metric.Desc().String()
		assert.Contains(desc, "variableLabels: {instance,player", "Metric description should contain the correct instance label")
	}
}

func TestCollectorDescribe(t *testing.T) {
	tMatrix := map[string]bool{
		"ReducedMetrics": false,
		"FullMetrics":    true,
	}
	for name, reduceMetrics := range tMatrix {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			c, err := NewSaveCollector("./testdata/1.20", "test-instance", reduceMetrics)
			require.NoError(err)

			expectedDescCount := 19

			ch := make(chan *prometheus.Desc)
			expectedDescs := make([]*prometheus.Desc, 0, expectedDescCount)
			go func() {
				prometheus.DescribeByCollect(c, ch)
				close(ch)
			}()
			for desc := range ch {
				expectedDescs = append(expectedDescs, desc)
			}

			ch = make(chan *prometheus.Desc)
			result := make([]*prometheus.Desc, 0, expectedDescCount)
			go func() {
				c.Describe(ch)
				close(ch)
			}()
			for desc := range ch {
				result = append(result, desc)
			}

			assert.Len(result, expectedDescCount, "Should have correct number of described metrics")
			for _, desc := range expectedDescs {
				assert.Contains(result, desc, "Described metrics should match collected metrics")
			}
		})
	}

}
