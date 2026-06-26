package sandbox

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	RPCCodeNotFound      = "not_found"
	RPCCodeAlreadyExists = "already_exists"
)

type RPCError struct {
	Code    string
	Message string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error [%s]: %s", e.Code, e.Message)
}

type RPC struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

func NewRPC(baseURL string, httpClient *http.Client, headers map[string]string) *RPC {
	return &RPC{baseURL: baseURL, httpClient: httpClient, headers: headers}
}

func (c *RPC) CallUnary(ctx context.Context, service, method string, req, resp any, extraHeaders ...map[string]string) error {
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, service, method)
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("connect: marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("connect: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Connect-Protocol-Version", "1")
	c.setHeaders(httpReq)
	for _, h := range extraHeaders {
		for k, v := range h {
			httpReq.Header.Set(k, v)
		}
	}
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("connect: request failed: %w", err)
	}
	defer httpResp.Body.Close()
	respBody, _ := io.ReadAll(httpResp.Body)
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("connect: POST %s: %w", url, parseRPCError(httpResp.StatusCode, respBody))
	}
	if resp != nil && len(respBody) > 0 {
		return json.Unmarshal(respBody, resp)
	}
	return nil
}

func (c *RPC) CallServerStream(ctx context.Context, service, method string, req any, timeoutMs int, extraHeaders ...map[string]string) (*RPCStream, error) {
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, service, method)
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("connect: marshal request: %w", err)
	}
	envelopedBody := encodeEnvelope(jsonBody)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(envelopedBody))
	if err != nil {
		return nil, fmt.Errorf("connect: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/connect+json")
	httpReq.Header.Set("Connect-Protocol-Version", "1")
	if timeoutMs > 0 {
		httpReq.Header.Set("Connect-Timeout-Ms", strconv.Itoa(timeoutMs))
	}
	c.setHeaders(httpReq)
	for _, h := range extraHeaders {
		for k, v := range h {
			httpReq.Header.Set(k, v)
		}
	}
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("connect: request failed: %w", err)
	}
	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close()
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("connect: POST %s: %w", url, parseRPCError(httpResp.StatusCode, respBody))
	}
	return &RPCStream{reader: httpResp.Body}, nil
}

func (c *RPC) setHeaders(req *http.Request) {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
}

func encodeEnvelope(data []byte) []byte {
	header := make([]byte, 5)
	header[0] = 0
	binary.BigEndian.PutUint32(header[1:5], uint32(len(data)))
	return append(header, data...)
}

type RPCStream struct {
	reader io.ReadCloser
}

func (s *RPCStream) Next(msg any) error {
	header := make([]byte, 5)
	if _, err := io.ReadFull(s.reader, header); err != nil {
		return err
	}
	flags := header[0]
	dataLen := binary.BigEndian.Uint32(header[1:5])
	data := make([]byte, dataLen)
	if _, err := io.ReadFull(s.reader, data); err != nil {
		return fmt.Errorf("connect: read body: %w", err)
	}
	if flags&0x02 != 0 {
		var trailer struct {
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		}
		if err := json.Unmarshal(data, &trailer); err == nil && trailer.Error != nil {
			return &RPCError{Code: trailer.Error.Code, Message: trailer.Error.Message}
		}
		return io.EOF
	}
	return json.Unmarshal(data, msg)
}

func (s *RPCStream) Close() error { return s.reader.Close() }

func parseRPCError(statusCode int, body []byte) error {
	var errResp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Code != "" {
		return &RPCError{Code: errResp.Code, Message: errResp.Message}
	}
	return fmt.Errorf("HTTP %d: %s", statusCode, string(body))
}
