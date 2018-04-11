package main

//Msg is JSON Data
type Msg struct {
	Consumer  string `json:"consumer"`
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Value     string `json:"value"`
}
