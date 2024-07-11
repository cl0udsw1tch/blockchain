package client

import (
	"sync"
)
func MakeClient() ClientId {
	var id ClientId
	id.MakeClientId()
	return id

}


func MakeClients(count int) []ClientId {
	clients := make([]ClientId, count)
	wg := sync.WaitGroup{};
	for i := range(count){
		wg.Add(1)
		go func () {
			defer wg.Done()
			clients[i] = MakeClient()
		}()
	}
	wg.Wait()
	return clients
}


