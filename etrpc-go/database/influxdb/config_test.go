// Package influxdb provides influxdb v1 client
package influxdb

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseAddress(t *testing.T) {
	testCases := []struct {
		name       string
		argAddress string
		wantRes    *UserConfig
		wantErr    error
	}{
		{
			name:       "influxdb",
			argAddress: "influxdb://user:password@ip:port?timeout=1000",
			wantRes: &UserConfig{
				Address:  "ip:port",
				Username: "user",
				Password: "password",
				Timeout:  time.Second,
			},
			wantErr: nil,
		},
		{
			name:       "influxdb+polaris",
			argAddress: "influxdb+polaris://user:password@trpc.influxdb.xxx.xxx?timeout=1",
			wantRes: &UserConfig{
				Address:  "trpc.influxdb.xxx.xxx",
				Username: "user",
				Password: "password",
				Timeout:  time.Millisecond,
			},
			wantErr: nil,
		},
		{
			name:       "empty",
			argAddress: "user:password@trpc.influxdb.xxx.xxx?timeout=1",
			wantRes: &UserConfig{
				Address:  "trpc.influxdb.xxx.xxx",
				Username: "user",
				Password: "password",
				Timeout:  time.Millisecond,
			},
			wantErr: nil,
		},
		{
			name:       "no timeout",
			argAddress: "user:password@trpc.influxdb.xxx.xxx",
			wantRes: &UserConfig{
				Address:  "trpc.influxdb.xxx.xxx",
				Username: "user",
				Password: "password",
				Timeout:  time.Second * 10,
			},
			wantErr: nil,
		},
		{
			name:       "no password",
			argAddress: "influxdb://user@trpc.influxdb.xxx.xxx",
			wantRes:    nil,
			wantErr:    fmt.Errorf("address:%s format invalid: user and password", "influxdb://user@trpc.influxdb.xxx.xxx"),
		},
		{
			name:       "no address",
			argAddress: "influxdb://user:password",
			wantRes:    nil,
			wantErr:    fmt.Errorf("address:%s format invalid: addr", "influxdb://user:password"),
		},
		{
			name:       "no address2",
			argAddress: "influxdb://user:password@",
			wantRes:    nil,
			wantErr:    fmt.Errorf("address:%s format invalid: addr", "influxdb://user:password@"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := ParseAddress(tc.argAddress)
			assert.EqualValues(t, tc.wantRes, res)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
