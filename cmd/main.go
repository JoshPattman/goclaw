package main

func main() {
	root := "/goclaw-data"
	err := UpdateConfig(root)
	if err != nil {
		panic(err)
	}
	data, err := LoadData(root)
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
