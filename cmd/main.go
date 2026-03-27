package main

import "flag"

func main() {
	root := flag.String("root", "/goclaw-data", "The root folder for data to be stored at")
	flag.Parse()

	err := UpdateConfig(*root)
	if err != nil {
		panic(err)
	}
	data, err := LoadData(*root)
	if err != nil {
		panic(err)
	}
	ag, err := CreateAgent(data)
	if err != nil {
		panic(err)
	}
	err = ag.Run()
	if err != nil {
		panic(err)
	}
}
