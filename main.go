package main

import (
	"d1agnoze/jsonimport/internal/server"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

// stdioConn adapts stdin/stdout to io.ReadWriteCloser
type stdioConn struct {
	in  *os.File
	out *os.File
}

func (c *stdioConn) Read(p []byte) (n int, err error) {
	return c.in.Read(p)
}

func (c *stdioConn) Write(p []byte) (n int, err error) {
	return c.out.Write(p)
}

func (c *stdioConn) Close() error {
	// Don’t actually close stdin/stdout, just return nil
	return nil
}

func main() {
	server := server.New()
	rpc.RegisterName("Server", &server)
	codec := jsonrpc.NewServerCodec(&stdioConn{os.Stdin, os.Stdout})
	print("starting server")
	rpc.ServeCodec(codec)
}
