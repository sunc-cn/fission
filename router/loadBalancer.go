package main

import "fmt"

type Server struct {
	id string
	weight int64
	currentWeight int64
}

type LoadBalancer struct {
	Servers map[string]*Server
}

func(lb *LoadBalancer) addServer(server *Server) {
	lb.Servers[server.id] = server
}

func(lb *LoadBalancer) getServer() *Server{
	var bestPeer *Server

	for _, v := range lb.Servers {

		v.currentWeight += v.weight

		if bestPeer == nil || bestPeer.currentWeight < v.currentWeight {
			bestPeer = v
			fmt.Printf("bestPeer: %s\n", v.id)
		}

	}

	if bestPeer != nil {
		bestPeer.currentWeight -= 100
	}

	lb.DumpServers()

	return bestPeer
}

func(lb *LoadBalancer) DumpServers() {
	fmt.Printf("Printing server info\n")

	for k, v := range lb.Servers {
		fmt.Printf("Server : %s\n", k)
		fmt.Printf("Contents : %v\n", *v)
		fmt.Printf("\n\n")
	}

	fmt.Printf("***********************************\n")
}

func main() {

	a := &Server{
		id: "a",
		weight: 80,
		currentWeight: 80,
	}

	b := &Server{
		id: "b",
		weight: 10,
		currentWeight: 10,
	}

	c := &Server{
		id: "c",
		weight: 10,
		currentWeight: 10,
	}

	lb := &LoadBalancer{
		Servers: make(map[string]*Server, 3),
	}

	lb.addServer(a)
	lb.addServer(b)
	lb.addServer(c)

	for i := 0 ; i < 10; i++ {
		fmt.Printf("Request %d\n", i)
		server := lb.getServer()
		fmt.Printf("Result Server : %v\n", *server)
		fmt.Printf("***************************************\n\n")
	}

}