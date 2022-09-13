package main

func main() {
	sever := NewSever("127.0.0.1", 9999)
	sever.Start()
}
