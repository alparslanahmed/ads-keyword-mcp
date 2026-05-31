package mcp

import (
	"bufio"
	"bytes"
	"context"
	"io"
)

// maxMessageBytes bounds a single JSON-RPC line.
const maxMessageBytes = 16 * 1024 * 1024

// StdioTransport serves the MCP server over newline-delimited JSON-RPC on
// stdin/stdout. Nothing is written to stdout except JSON-RPC responses; all logs
// go to stderr via the server's logger.
type StdioTransport struct {
	server *Server
	in     io.Reader
	out    io.Writer
}

// NewStdioTransport builds a stdio transport.
func NewStdioTransport(server *Server, in io.Reader, out io.Writer) *StdioTransport {
	return &StdioTransport{server: server, in: in, out: out}
}

// Serve reads messages until EOF or context cancellation.
func (t *StdioTransport) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(t.in)
	scanner.Buffer(make([]byte, 0, 64*1024), maxMessageBytes)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		// Copy the line: the scanner reuses its buffer on the next Scan.
		msg := make([]byte, len(line))
		copy(msg, line)

		resp := t.server.HandleMessage(ctx, msg)
		if resp == nil {
			continue
		}
		if err := t.write(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (t *StdioTransport) write(resp []byte) error {
	if _, err := t.out.Write(resp); err != nil {
		return err
	}
	_, err := t.out.Write([]byte("\n"))
	return err
}
