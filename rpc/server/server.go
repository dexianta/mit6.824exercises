package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
	dxrpc "dexian.io/mit/distsys/rpc"
)


type Master struct {
	tasks []dxrpc.Task
}

func (m *Master) GetTask(args dxrpc.Args, reply *dxrpc.Reply) error {
	fmt.Println("by default, reply is: ", reply.Task)
	fmt.Printf("worker asking for task: %v\n", args.FinishedTask)
	for i := 0; i < len(m.tasks); i++ {
		e := &m.tasks[i]
		if e.TaskName == args.FinishedTask.TaskName {
			m.tasks[i].Done = true
		}
		// fmt.Println("---: ", e)
	}

	for _, elem := range m.tasks {
		if !elem.Done {
			fmt.Println("we are sending: ", elem)
			reply.Task = &elem
			break
		}
	}


	return nil
}


func server() {
	master := new (Master)
	master.tasks = []dxrpc.Task{}
	master.tasks = []dxrpc.Task{
		dxrpc.Task{TaskName: "task1"},
		dxrpc.Task{TaskName: "task2"},
		dxrpc.Task{TaskName: "task3"},
	}

	// rpcs := rpc.NewServer()
	rpc.Register(master)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":4321")
	if e != nil {
		log.Fatal("listen error: ", e.Error())
	}
	go http.Serve(l, nil)
}

func main() {
	server()
	fmt.Println("server started")
	for {
		time.Sleep(time.Minute)
	}

	// t := dxrpc.Task{TaskName: "none", Done: false}
	// t.TaskName = "haha"
	// t.Done = true
	// fmt.Println(t)
	time.Sleep(time.Second)
}

