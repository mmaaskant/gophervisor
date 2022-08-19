# Gophervisor
> A simple, lightweight concurrency manager in Go.

Gophervisor runs and monitors workers that can handle simple tasks concurrently.

## Features

By creating a new instance of the Supervisor class you can do the following:
```go
package main

import (
	"fmt"
	"github.com/mmaaskant/gophervisor/pkg/supervisor"
)

func main() {
	sv := supervisor.NewSupervisor(10) // Create a new supervisor with 10 workers in its pool
	p, rch := sv.Register(func(p *supervisor.Publisher, d interface{}, rch chan interface{}) {
		fmt.Println(d)                  // A unit of data that has been published
		p.Publish("A new unit of data") // Publish a new unit of data to be processed using the same registered function
		rch <- "A response message"     // Send a response
	})
	p.Publish("Unit of data")   // Publish a new unit of data to the worker pool to process
	go sv.Shutdown()            // Start the graceful shutdown of the Supervisor and skip waiting for the shutdown to finish
	for response := range rch { // Get responses, due to shutdown loop will end once all workers have completed
		fmt.Println(response) //  // A single response from the worker pool
	}
}
```


## Licensing

The code in this project is licensed under an GPL-3.0 license.