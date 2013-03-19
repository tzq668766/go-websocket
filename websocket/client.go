// Copyright 2013 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package websocket

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// NewClient creates a new client connection using the given net connection.
// The URL u specifies the host and request URI. The header specifies optional
// Origin, Sec-WebSocket-Protocol and Cookie headers.
func NewClient(netConn net.Conn, u *url.URL, header http.Header, readBufSize, writeBufSize int) (c *Conn, subprotocol string, err error) {
	challengeKey, err := generateChallengeKey()
	if err != nil {
		return nil, "", err
	}
	acceptKey := computeAcceptKey(challengeKey)

	c = newConn(netConn, false, readBufSize, writeBufSize)
	p := c.writeBuf[:0]
	p = append(p, "GET "...)
	p = append(p, u.RequestURI()...)
	p = append(p, " HTTP/1.1\r\nHost: "...)
	p = append(p, u.Host...)
	p = append(p, "\r\nUpgrade: websocket\r\nConnection: upgrade\r\nSec-WebSocketVersion: 13\r\nSec-WebSocket-Key: "...)
	p = append(p, challengeKey...)
	p = append(p, "\r\n"...)
	for k, vs := range header {
		for _, v := range vs {
			p = append(p, k...)
			p = append(p, ": "...)
			p = append(p, v...)
			p = append(p, "\r\n"...)
		}
	}
	p = append(p, "\r\n"...)

	if _, err := netConn.Write(p); err != nil {
		return nil, "", err
	}

	resp, err := http.ReadResponse(c.br, &http.Request{Method: "GET"})
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != 101 ||
		strings.ToLower(resp.Header.Get("Upgrade")) != "websocket" ||
		strings.ToLower(resp.Header.Get("Connection")) != "upgrade" ||
		resp.Header.Get("Sec-WebSocket-Accept") != acceptKey {
		return nil, "", errors.New("websocket: bad handshake")
	}
	return c, resp.Header.Get("Sec-WebSocket-Protocol"), nil
}
