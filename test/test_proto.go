package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/json-iterator/go"

	//"github.com/json-iterator/go"
	pb "github.com/vearne/chat/proto"
)

func main() {
	temp := &pb.PushSignal{SenderId: 15,
		Data: &pb.PushSignal_Partner{&pb.AccountInfo{AccountId: 15, NickName: "xxxx"}}}
	data, err := proto.Marshal(temp)
	if err != nil {
		fmt.Println(err)
	}
	target := &pb.PushSignal{}
	proto.Unmarshal(data, target)
	fmt.Println(jsoniter.MarshalToString(target))
}
