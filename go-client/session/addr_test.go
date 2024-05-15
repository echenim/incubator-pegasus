/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package session

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveMetaAddr(t *testing.T) {
	addrs := []string{"127.0.0.1:34601", "127.0.0.1:34602", "127.0.0.1:34603"}
	resolvedAddrs, err := ResolveMetaAddr(addrs)
	assert.Nil(t, err)
	assert.Equal(t, addrs, resolvedAddrs)

	addrs = []string{"127.0.0.1:34601", "www.baidu.com", "127.0.0.1:34603"}
	_, err = ResolveMetaAddr(addrs)
	assert.NotNil(t, err)

	addrs = []string{"www.baidu.com"}
	_, err = ResolveMetaAddr(addrs)
	assert.Nil(t, err)
	assert.Greater(t, len(addrs), 0)

	addrs = []string{"abcde"}
	_, err = ResolveMetaAddr(addrs)
	assert.NotNil(t, err)

	addrs = nil
	_, err = ResolveMetaAddr(addrs)
	assert.NotNil(t, err)

	addrs = []string{}
	_, err = ResolveMetaAddr(addrs)
	assert.NotNil(t, err)
}

// MockResolver allows us to simulate DNS and TCP address resolution
type MockResolver struct {
	tcpAddrs  map[string]*net.TCPAddr
	hostAddrs map[string][]string
}

func (r *MockResolver) ResolveTCPAddr(network, address string) (*net.TCPAddr, error) {
	if addr, ok := r.tcpAddrs[address]; ok {
		return addr, nil
	}
	return nil, errors.New("mock: failed to resolve TCP address")
}

func (r *MockResolver) LookupHost(host string) ([]string, error) {
	if addrs, ok := r.hostAddrs[host]; ok {
		return addrs, nil
	}
	return nil, errors.New("mock: failed to lookup host")
}

func TestImprovedResolveMetaAddr(t *testing.T) {
	mockResolver := &MockResolver{
		tcpAddrs: map[string]*net.TCPAddr{
			"127.0.0.1:34601": {IP: net.ParseIP("127.0.0.1"), Port: 34601},
			"127.0.0.1:34602": {IP: net.ParseIP("127.0.0.1"), Port: 34602},
			"127.0.0.1:34603": {IP: net.ParseIP("127.0.0.1"), Port: 34603},
		},
		hostAddrs: map[string][]string{
			"www.baidu.com": {"192.168.1.1", "192.168.1.2"},
		},
	}

	tests := []struct {
		name       string
		addrs      []string
		wantErr    bool
		wantLength int
	}{
		{
			name:       "All TCP addresses",
			addrs:      []string{"127.0.0.1:34601", "127.0.0.1:34602", "127.0.0.1:34603"},
			wantErr:    false,
			wantLength: 3,
		},
		{
			name:       "Mixed TCP and hostname addresses",
			addrs:      []string{"127.0.0.1:34601", "www.baidu.com", "127.0.0.1:34603"},
			wantErr:    false,
			wantLength: 5, // 3 TCP addresses + 2 resolved from www.baidu.com
		},
		{
			name:       "Single hostname address",
			addrs:      []string{"www.baidu.com"},
			wantErr:    false,
			wantLength: 2, // 2 resolved from www.baidu.com
		},
		{
			name:       "Invalid address",
			addrs:      []string{"abcde"},
			wantErr:    true,
			wantLength: 0,
		},
		{
			name:       "Nil addresses",
			addrs:      nil,
			wantErr:    true,
			wantLength: 0,
		},
		{
			name:       "Empty addresses",
			addrs:      []string{},
			wantErr:    true,
			wantLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolvedAddrs, err := ResolveMetaAddr(tt.addrs, mockResolver)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantLength, len(resolvedAddrs))
			}
		})
	}
}
