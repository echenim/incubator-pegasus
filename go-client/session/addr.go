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
	"fmt"
	"net"
)

// Resolver interface allows mocking net functions
type Resolver interface {
	ResolveTCPAddr(network, address string) (*net.TCPAddr, error)
	LookupHost(host string) ([]string, error)
}

// DefaultResolver uses the net package functions
type DefaultResolver struct{}

func (r *DefaultResolver) ResolveTCPAddr(network, address string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr(network, address)
}

func (r *DefaultResolver) LookupHost(host string) ([]string, error) {
	return net.LookupHost(host)
}

// ResolveMetaAddr into a list of TCP4 addresses. Error is returned if the given `addrs` are not either
// a list of valid TCP4 addresses, or a resolvable hostname.
func ResolveMetaAddr(addrs []string) ([]string, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("meta server list should not be empty")
	}

	var resolvedAddrs []string
	var invalidAddrs []string

	// Loop through each address in the input list
	for _, addr := range addrs {
		// Try to resolve the address as a TCP4 address
		tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
		if err == nil {
			// If successful, append the resolved TCP address to the list
			resolvedAddrs = append(resolvedAddrs, tcpAddr.String())
			continue
		}

		// If TCP4 resolution fails, try to resolve the address as a hostname
		hostAddrs, err := net.LookupHost(addr)
		if err == nil {
			for _, hostAddr := range hostAddrs {
				// Resolve each resolved hostname IP as a TCP4 address
				tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:0", hostAddr))
				if err == nil {
					// Append the resolved IP address to the list
					resolvedAddrs = append(resolvedAddrs, tcpAddr.IP.String())
				}
			}
			continue
		}

		// If both TCP4 and hostname resolution fail, add the address to the list of invalid addresses
		invalidAddrs = append(invalidAddrs, addr)
	}

	// If no addresses could be resolved, return an error with the list of invalid addresses
	if len(resolvedAddrs) == 0 {
		return nil, fmt.Errorf("no valid TCP4 addresses or resolvable hostnames found: %v", invalidAddrs)
	}

	// Return the list of resolved addresses
	return resolvedAddrs, nil
}
