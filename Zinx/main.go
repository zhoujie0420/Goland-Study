package main

import "Zinx/znet"

func main() {
	s := znet.NewServer("[zinx]")
	s.Serve()
}
