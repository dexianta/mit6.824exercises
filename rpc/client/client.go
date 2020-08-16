package main

import (
	"log"
	"net/rpc"
	"time"
	"fmt"
	dxrpc "dexian.io/mit/distsys/rpc"
)

func main()  {
	client, err := rpc.DialHTTP("tcp", ":4321")
	if err != nil {
		log.Fatal("dialing: ", err.Error())
	}
	task := dxrpc.Task{TaskName: "none", Done: false}

	for {
		var reply dxrpc.Reply
		err = client.Call("Master.GetTask", dxrpc.Args{FinishedTask: task}, &reply)

		if err != nil {
			log.Fatal("can't get task: ", err.Error())
		}
		fmt.Printf("%v\n", reply.Task)
		task.TaskName = reply.Task.TaskName
		task.Done = true
		time.Sleep(5 * time.Second)
	}
}