package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
	"github.com/tiereum/trmclient/internal/t_config"
	"github.com/tiereum/trmclient/internal/t_error"
)

type Relayer struct {
	publicEndpoint string
}



func getRelayEndpoint() string {
	return "http://34.16.104.117:8033"
}


type Network struct {
	ctx     *t_config.Context
	istream chan []byte
	ostream chan []byte
	relayer *Relayer
	port	*uint16
}

func NewNetwork(ctx *t_config.Context) *Network {

	network := new(Network)
	network.ctx = ctx
	network.istream = make(chan []byte)
	network.ostream = make(chan []byte)

	network.relayer = new(Relayer)
	network.port = ctx.NetworkConfig.ClientPort
	network.relayer.publicEndpoint = getRelayEndpoint()

	return network
}

// to replaced by P2P module
func (network *Network) Broadcast(data []byte, inputChannel chan []byte, outputChan chan []byte) {

	localAddr := fmt.Sprintf(`::{%d}`, &network.ctx.NetworkConfig.ClientPort) // Bind to port 12345 locally (you can change this)
	remoteAddr := network.relayer.publicEndpoint// The remote server you want to dial

	// Resolve the local address
	local, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		log.Fatalf("Failed to resolve local address: %v", err)
	}

	// Resolve the remote address
	remote, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		log.Fatalf("Failed to resolve remote address: %v", err)
	}

	// Create a Dialer with the local address bound
	dialer := &net.Dialer{
		LocalAddr: local, // Bind to the local address
		Timeout:   10 * time.Second, // Dial timeout
	}

	conn, err := dialer.Dial("tcp", remote.String())
	t_error.LogErr(err)

	defer conn.Close()
	fmt.Println("Listening on all interfaces on port " + fmt.Sprint(network.ctx.NetworkConfig.ClientPort) + "\n")



	go network.handleConn(conn, data)
	

}


type ClientResponse struct {
	Code int8
	Payload []byte
}

func (network *Network) handleConn(conn net.Conn, data []byte) *ClientResponse {

	fmt.Print("Sending data ...")
	_, err := conn.Write(data)
	t_error.LogErr(err)

	b := []byte{}
	fmt.Print("Receiving data ...")
	n, err := conn.Read(b)
	t_error.LogErr(err)

	res := ClientResponse{}
	json.Unmarshal(b[:n], &res)

	return &res

}


func (network *Network) READ() {
	
}

func (network *Network) UPDATE() {
	
}

func (network *Network) DELETE() {
	
}
