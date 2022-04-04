package qos_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kolotaev/qos"
)

func TestThrottler_NewThrottler(t *testing.T) {
	th := qos.NewThrottler(50, false)
	assert.NotNil(t, th)
	assert.False(t, th.IsEnabled())
}

func TestThrottler_Enabling(t *testing.T) {
	th := qos.NewThrottler(50, true)
	assert.True(t, th.IsEnabled())
	th.Disable()
	assert.False(t, th.IsEnabled())
	th.Enable()
	assert.True(t, th.IsEnabled())
}

func TestThrottler_ChangeBandwidthLimits(t *testing.T) {
	th := qos.NewThrottler(50, true)
	th.SetBandwidthLimitForConnection(40, "abc")
	assert.Equal(t, int64(40), th.GetBandwidthLimitForConnection("abc"))
	assert.Equal(t, int64(50), th.GetBandwidthLimitForConnection("xyz"))
	th.SetBandwidthLimit(100)
	assert.Equal(t, int64(100), th.GetBandwidthLimitForConnection("abc"))
	assert.Equal(t, int64(100), th.GetBandwidthLimitForConnection("xyz"))
}

func TestThrottler_UnregisterConnection(t *testing.T) {
	th := qos.NewThrottler(50, true)
	th.SetBandwidthLimitForConnection(40, "abc")
	assert.Equal(t, int64(40), th.GetBandwidthLimitForConnection("abc"))
	th.UnregisterConnection("abc")
	assert.Equal(t, int64(50), th.GetBandwidthLimitForConnection("abc"))
}

func TestThrottler_WriteWithDisabled(t *testing.T) {
	th := qos.NewThrottler(1, false)
	out := bytes.NewBufferString("")
	ctx := context.Background()
	contents := strings.NewReader("Address tradeoff between development cycle time and server performance")
	n, err := th.Write(ctx, out, "abc", contents)
	assert.NoError(t, err)
	assert.Equal(t, int64(70), n)
	assert.Equal(t, "Address tradeoff between development cycle time and server performance", out.String())
}

func TestThrottler_WriteWithTimeout(t *testing.T) {
	th := qos.NewThrottler(4, true)
	out := bytes.NewBufferString("")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	contents := strings.NewReader("Address tradeoff between development cycle time and server performance")
	n, err := th.Write(ctx, out, "abc", contents)
	assert.ErrorContains(t, err, "Wait(n=1) would exceed context deadline")
	assert.Equal(t, int64(4), n)
	assert.Equal(t, "Addr", out.String())
}
