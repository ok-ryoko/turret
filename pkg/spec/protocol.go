// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"fmt"
	"strings"
)

const (
	TCP Protocol = 1 << iota
	UDP
)

// Protocol is a unique identifier for a network protocol. The zero value
// represents an unknown protocol.
type Protocol uint

// String returns a string containing the stylized name of the protocol.
func (p Protocol) String() string {
	var s string
	switch p {
	case TCP:
		s = "tcp"
	case UDP:
		s = "udp"
	default:
		s = "unknown"
	}
	return s
}

// ProtocolWrapper wraps Protocol to facilitate its parsing from serialized data.
type ProtocolWrapper struct {
	Protocol
}

// UnmarshalText decodes the protocol from a string.
func (w *ProtocolWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Protocol, err = parseProtocolString(string(text))
	return err
}

func parseProtocolString(s string) (Protocol, error) {
	var p Protocol
	switch strings.ToLower(s) {
	case "tcp":
		p = TCP
	case "udp":
		p = UDP
	default:
		return 0, fmt.Errorf("unsupported network protocol %q", s)
	}
	return p, nil
}
