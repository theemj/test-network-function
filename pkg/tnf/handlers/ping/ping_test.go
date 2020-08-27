package ping_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/ping"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"strconv"
	"testing"
	"time"
)

const (
	testDataDirectory   = "testdata"
	testDataFileSuffix  = ".txt"
	testTimeoutDuration = time.Second * 2
)

type TestCase struct {
	host             string
	count            int
	expectedSent     int
	expectedReceived int
	expectedErrors   int
	expectedResult   int
}

var testCases = map[string]TestCase{
	"ip_address_no_packet_loss": {
		host:             "192.168.1.1",
		count:            4,
		expectedSent:     4,
		expectedReceived: 4,
		expectedErrors:   0,
		expectedResult:   tnf.SUCCESS,
	},
	"hostname_no_packet_loss": {
		host:             "www.google.com",
		count:            10,
		expectedSent:     10,
		expectedReceived: 10,
		expectedErrors:   0,
		expectedResult:   tnf.SUCCESS,
	},
	"ip_address_error_packet_loss": {
		host:             "192.168.1.1",
		count:            20,
		expectedSent:     20,
		expectedReceived: 16,
		expectedErrors:   4,
		expectedResult:   tnf.ERROR,
	},
	"ip_address_failing_packet_loss": {
		host:             "192.168.1.2",
		count:            1,
		expectedSent:     1,
		expectedReceived: 0,
		expectedErrors:   0,
		expectedResult:   tnf.FAILURE,
	},
	"ip_address_passing_packet_loss": {
		host:             "192.168.1.1",
		count:            20,
		expectedSent:     20,
		expectedReceived: 19,
		expectedErrors:   0,
		expectedResult:   tnf.SUCCESS,
	},
	"incorrect_ip_address": {
		host:             "0.0.1.2",
		count:            1,
		expectedSent:     0,
		expectedReceived: 0,
		expectedErrors:   0,
		expectedResult:   tnf.ERROR,
	},
}

func getMockOutputFilename(testName string) string {
	return path.Join(testDataDirectory, fmt.Sprintf("%s%s", testName, testDataFileSuffix))
}

func getMockOutput(t *testing.T, testName string) string {
	fileName := getMockOutputFilename(testName)
	b, err := ioutil.ReadFile(fileName)
	assert.Nil(t, err)
	return string(b)
}

func TestNewPing(t *testing.T) {
	for _, testCase := range testCases {
		request := ping.NewPing(testTimeoutDuration, testCase.host, testCase.count)
		assert.NotNil(t, request)
		args := ping.PingCmd(testCase.host, testCase.count)
		assert.Equal(t, args, request.Args())
		assert.Equal(t, tnf.ERROR, request.Result())
	}
}

func TestPing_Args(t *testing.T) {
	for _, testCase := range testCases {
		request := ping.NewPing(testTimeoutDuration, testCase.host, testCase.count)
		assert.NotNil(t, request)
		args := []string{"ping", "-c", strconv.Itoa(testCase.count), testCase.host}
		assert.Equal(t, args, request.Args())
	}
}

func TestPing_ReelFirst(t *testing.T) {
	for _, testCase := range testCases {
		request := ping.NewPing(testTimeoutDuration, testCase.host, testCase.count)
		step := request.ReelFirst()
		assert.Equal(t, "", step.Execute)
		assert.NotNil(t, step.Expect)
		for _, expectation := range step.Expect {
			assert.Contains(t, request.GetReelFirstRegularExpressions(), expectation)
		}
		assert.Equal(t, testTimeoutDuration, step.Timeout)
	}
}

func TestPing_GetStats(t *testing.T) {
	request := ping.NewPing(testTimeoutDuration, "192.168.1.1", 1)
	sent, received, errors := request.GetStats()
	assert.Zero(t, sent)
	assert.Zero(t, received)
	assert.Zero(t, errors)
}

func TestPing_ReelMatch(t *testing.T) {
	for testCaseName, testCase := range testCases {
		request := ping.NewPing(testTimeoutDuration, testCase.host, testCase.count)
		matchMock := getMockOutput(t, testCaseName)
		step := request.ReelMatch("", "", matchMock)
		assert.Nil(t, step)
		actualSent, actualReceived, actualErrors := request.GetStats()
		assert.Equal(t, testCase.expectedSent, actualSent)
		assert.Equal(t, testCase.expectedReceived, actualReceived)
		assert.Equal(t, testCase.expectedErrors, actualErrors)
		actualResult := request.Result()
		assert.Equal(t, testCase.expectedResult, actualResult)
	}
}

func TestPing_ReelTimeout(t *testing.T) {
	request := ping.NewPing(testTimeoutDuration, "192.168.1.2", 1)
	step := request.ReelTimeout()
	assert.NotNil(t, step)
	assert.NotNil(t, step.Execute)
	var expect []string
	assert.Equal(t, expect, step.Expect)
}

func TestPing_Timeout(t *testing.T) {
	for _, testCase := range testCases {
		request := ping.NewPing(testTimeoutDuration, testCase.host, testCase.count)
		assert.Equal(t, testTimeoutDuration, request.Timeout())
	}
}

// Just ensure there are no panics.
func TestPing_ReelEof(t *testing.T) {
	for _, testCase := range testCases {
		request := ping.NewPing(testTimeoutDuration, testCase.host, testCase.count)
		request.ReelEof()
	}
}

func TestPingCmd(t *testing.T) {
	cmd := ping.PingCmd("192.168.1.1", 0)
	assert.Equal(t, []string{"ping", "192.168.1.1"}, cmd)
	cmd = ping.PingCmd("192.168.1.1", -1)
	assert.Equal(t, []string{"ping", "192.168.1.1"}, cmd)
	cmd = ping.PingCmd("192.168.1.1", 1)
	assert.Equal(t, []string{"ping", "-c", "1", "192.168.1.1"}, cmd)
}
