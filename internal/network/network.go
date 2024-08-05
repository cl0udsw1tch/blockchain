package network

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
)

type Network struct {
	ctx *t_config.Context
	istream chan []byte
	ostream chan []byte

}

func NewNetwork(ctx *t_config.Context) *Network {
	
	network := new(Network)
	network.ctx = ctx
	network.istream = make(chan []byte)
	network.ostream = make(chan []byte)
	
	tablePath := path.Join(ctx.TmpDir, "nodeTable.txt")

	f, err := os.OpenFile(tablePath, os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0666)
	t_error.LogErr(err)
	defer f.Close()
	f.WriteString(fmt.Sprint(ctx.NodeConfig.RpcEndpointPort) + "\n")
	
	return network
}

// to replaced by P2P module
func (network *Network) Listen() {

	listener, err := net.Listen("tcp", ":" + fmt.Sprint(network.ctx.NodeConfig.RpcEndpointPort))
	t_error.LogErr(err)

	defer listener.Close()
	fmt.Println("Server listening on port " +  fmt.Sprint(network.ctx.NodeConfig.RpcEndpointPort))

	for {
		conn, err := listener.Accept()
		t_error.LogErr(err)
		go network.handleConn(conn)
	}
	
	
}

func (network *Network) handleConn(conn net.Conn) {

	fmt.Println("New client connected: ", conn.RemoteAddr().String())

	var b []byte
	_, err := conn.Read(b)
	t_error.LogErr(err)
	network.istream <- b

}

// to replaced by P2P module
func (network *Network) Broadcast() {

	portPath := path.Join(network.ctx.TmpDir, "nodeTable.txt")
	f, err := os.Open(portPath)
	t_error.LogErr(err)
	defer f.Close()
	t_error.LogErr(err)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		port := scanner.Text()
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%s", port))
		t_error.LogErr(err)
		defer conn.Close()
	
		data := <-network.ostream
		_, err = conn.Write(data)
		t_error.LogErr(err)
	}
}

func (network *Network) IStream() <-chan []byte {
	return network.istream
}

func (network *Network) OStream() chan<- []byte {
	return network.ostream
}