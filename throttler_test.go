package qos_test

import (
	"bytes"
	"context"
	"net"
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

func TestThrottler_RegisterUnregisterConnectionWithoutSpecificLimits(t *testing.T) {
	th := qos.NewThrottler(50, true)
	th.RegisterConnection("A")
	assert.Equal(t, int64(50), th.GetBandwidthLimitForConnection("A"))
	th.UnregisterConnection("A")
	assert.Equal(t, int64(0), th.GetBandwidthLimitForConnection("A"))
	th.RegisterConnection("A")
	assert.Equal(t, int64(50), th.GetBandwidthLimitForConnection("A"))
}

func TestThrottler_RegisterUnregisterConnectionWithSpecificLimit(t *testing.T) {
	th := qos.NewThrottler(50, true)
	th.SetBandwidthLimitForConnection(20, "C")
	assert.Equal(t, int64(20), th.GetBandwidthLimitForConnection("C"))
	th.UnregisterConnection("C")
	assert.Equal(t, int64(0), th.GetBandwidthLimitForConnection("C"))
	th.RegisterConnection("C")
	assert.Equal(t, int64(20), th.GetBandwidthLimitForConnection("C"))
}

func TestThrottler_ConnectionLimitsCalculations(t *testing.T) {
	th := qos.NewThrottler(50, true)

	th.RegisterConnection("A")
	th.RegisterConnection("B")
	th.SetBandwidthLimitForConnection(20, "C")
	assert.Equal(t, int64(15), th.GetBandwidthLimitForConnection("A"))
	assert.Equal(t, int64(15), th.GetBandwidthLimitForConnection("B"))
	assert.Equal(t, int64(20), th.GetBandwidthLimitForConnection("C"))

	th.UnregisterConnection("C")
	assert.Equal(t, int64(25), th.GetBandwidthLimitForConnection("A"))
	assert.Equal(t, int64(25), th.GetBandwidthLimitForConnection("B"))

	th.SetBandwidthLimitForConnection(45, "D")
	assert.Equal(t, int64(45), th.GetBandwidthLimitForConnection("D"))
	assert.Equal(t, int64(2), th.GetBandwidthLimitForConnection("A"))
	assert.Equal(t, int64(2), th.GetBandwidthLimitForConnection("B"))
	th.UnregisterConnection("D")

	th.SetBandwidthLimitForConnection(30, "D")
	th.SetBandwidthLimitForConnection(10, "E")
	assert.Equal(t, int64(30), th.GetBandwidthLimitForConnection("D"))
	assert.Equal(t, int64(10), th.GetBandwidthLimitForConnection("E"))
	assert.Equal(t, int64(5), th.GetBandwidthLimitForConnection("A"))
	assert.Equal(t, int64(5), th.GetBandwidthLimitForConnection("B"))
	th.UnregisterConnection("D")
	th.UnregisterConnection("E")

	th.RegisterConnection("F")
	assert.Equal(t, int64(16), th.GetBandwidthLimitForConnection("A"))
	assert.Equal(t, int64(16), th.GetBandwidthLimitForConnection("B"))
	assert.Equal(t, int64(16), th.GetBandwidthLimitForConnection("F"))

	th.SetBandwidthLimitForConnection(30, "J")
	th.SetBandwidthLimitForConnection(40, "H") // exceeds max capacity
	assert.Equal(t, int64(30), th.GetBandwidthLimitForConnection("J"))
	assert.Equal(t, int64(20), th.GetBandwidthLimitForConnection("H"))
	assert.Equal(t, int64(0), th.GetBandwidthLimitForConnection("A"))
	assert.Equal(t, int64(0), th.GetBandwidthLimitForConnection("B"))
	assert.Equal(t, int64(0), th.GetBandwidthLimitForConnection("F"))
}

func TestThrottler_SetBandwidthLimitForConnectionThatExceedsFreePool(t *testing.T) {
	th := qos.NewThrottler(50, true)
	th.RegisterConnection("B")
	th.SetBandwidthLimitForConnection(60, "A")
	assert.Equal(t, int64(50), th.GetBandwidthLimitForConnection("A"))
	assert.Equal(t, int64(0), th.GetBandwidthLimitForConnection("B"))
}

func TestThrottler_ChangeBandwidthLimitsUpwards(t *testing.T) {
	th := qos.NewThrottler(50, true)
	th.RegisterConnection("xyz")
	th.SetBandwidthLimitForConnection(40, "abc")

	assert.Equal(t, int64(40), th.GetBandwidthLimitForConnection("abc"))
	assert.Equal(t, int64(10), th.GetBandwidthLimitForConnection("xyz"))

	th.SetBandwidthLimit(100)
	assert.Equal(t, int64(40), th.GetBandwidthLimitForConnection("abc"))
	assert.Equal(t, int64(60), th.GetBandwidthLimitForConnection("xyz"))
}

func TestThrottler_ChangeBandwidthLimitsDownwards(t *testing.T) {
	th := qos.NewThrottler(60, true)
	th.RegisterConnection("xyz")
	th.SetBandwidthLimitForConnection(30, "abc")
	th.SetBandwidthLimitForConnection(20, "qwe")

	assert.Equal(t, int64(30), th.GetBandwidthLimitForConnection("abc"))
	assert.Equal(t, int64(20), th.GetBandwidthLimitForConnection("qwe"))
	assert.Equal(t, int64(10), th.GetBandwidthLimitForConnection("xyz"))

	th.SetBandwidthLimit(5)
	assert.Equal(t, int64(2), th.GetBandwidthLimitForConnection("abc"))
	assert.Equal(t, int64(2), th.GetBandwidthLimitForConnection("abc"))
	assert.Equal(t, int64(0), th.GetBandwidthLimitForConnection("xyz"))
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

func TestThrottler_Listen(t *testing.T) {
	th := qos.NewThrottler(4, true)
	err := th.Listen("tcp", "127.0.0.1:0")
	defer th.Close()
	assert.NoError(t, err)
}

func TestThrottler_ListenOnWrongAddress(t *testing.T) {
	th := qos.NewThrottler(4, true)
	err := th.Listen("tcp", "127.0.0.1:999999999")
	assert.EqualError(t, err, "failed to listen with Throttler: listen tcp: address 999999999: invalid port")

	c, err := th.Accept()
	assert.Nil(t, c)
	assert.EqualError(t, err, "please start listening first")
}

func TestThrottler_ListenCanBeCalledOnlyOnce(t *testing.T) {
	th := qos.NewThrottler(4, true)
	err := th.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer th.Close()

	err = th.Listen("tcp", "127.0.0.1:0")
	assert.EqualError(t, err, "listening was started previously, it can be started only once")
}

func TestThrottler_AcceptBeforeListening(t *testing.T) {
	th := qos.NewThrottler(4, true)
	c, err := th.Accept()
	assert.Nil(t, c)
	assert.EqualError(t, err, "please start listening first")
}

func TestThrottler_Addr(t *testing.T) {
	th := qos.NewThrottler(4, true)
	err := th.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer th.Close()

	addr := th.Addr()
	assert.Contains(t, addr.String(), "127.0.0.1:")
}

func TestThrottler_AddrBeforeListening(t *testing.T) {
	th := qos.NewThrottler(4, true)
	addr := th.Addr()
	assert.Nil(t, addr)
}

func TestThrottler_NetListenerInterfaceCompilation(t *testing.T) {
	th := qos.NewThrottler(4, true)
	th.Listen("tcp", "127.0.0.1:0")
	defer th.Close()
	func(l net.Listener) {
		th.Addr()
	}(th)
	assert.NotNil(t, th)
}
