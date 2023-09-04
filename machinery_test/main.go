package main

import (
	"fmt"
	"log"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
)

func fatalErrorHandler(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	config, err := config.NewFromYaml("config.yml", false)
	fatalErrorHandler(err)
	server, err := machinery.NewServer(config)
	fatalErrorHandler(err)
	err = server.RegisterTask("sum", Sum)
	fatalErrorHandler(err)
	worker := server.NewWorker("xiwen", 2)
	go func() {
		err = worker.Launch()
		fatalErrorHandler(err)
	}()
	signature := &tasks.Signature{
		Name: "sum",
		Args: []tasks.Arg{
			{
				Type:"[]int64",
				Value: []int64{1, 2, 3, 4},
			},
		},
		RetryTimeout: 300,
		RetryCount: 3,
	}
	results, err := server.SendTask(signature)
	fatalErrorHandler(err)
	res, err := results.Get(1)
	fmt.Printf("the result is %v", tasks.HumanReadableResults(res))
}


func Sum(args []int64) (int64, error) {
	sum := int64(0)
	for _, arg := range args {
		sum += arg
	}
	return sum, nil
}
