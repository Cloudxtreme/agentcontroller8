# AgentController2 Client #
A client library for speaking to AgentController from Go.

# How it works #
AgentController2 acts mainly as a relay for administrative commands sent from clients to connected Agent2 instances, and for 
responses back to the client from the targeted Agent2 instances.
![Workflow](https://raw.githubusercontent.com/Jumpscale/agentcontroller2/master/newclient/ac2.png)

# Usage #
## 1. Choose your target(s) ##
```go
// Target any one node
target := commandfactory.CommandTarget{}
```
... or ...
```go
// Target all available nodes
target := commandfactory.CommandTarget{Fanout: true}
```
... or ...
```go
// Target any node on Grid 42
target := commandfactory.CommandTarget{GID: 42}
```
... or ...
```go
// Target all nodes on Grid 42
target := commandfactory.CommandTarget{GID: 42, Fanout: true}
```
... or ...
```go
// Target the specific node 23 on Grid 7
target := commandfactory.CommandTarget{GID: 7, NID: 23}
```

## 2. Issue commands to chosen targets ##

*Note:* Communication from a client to AgentController2 is achieved via passing messages back and forth on synchronized data structure in a shared Redis instance. In order to start communicating, you need to know the host name and port of your Redis instance used for communication.

You can use the high-level client's [rich non-blocking API](https://godoc.org/github.com/Jumpscale/agentcontroller2/newclient#Client) for issuing commands and receiving responses very easily.
```go
client := newclient.NewClient("localhost:9999", "")

target := newclient.AnyNode()

// For example, we'll command the target nodes to execute the "ls" executable with the arguments "/opt"
responseChan, errChan := client.ExecuteExecutable(target, "ls", []string{"/opt"})

// Since we're targeting a single node, we're expecting a single response
// If we were targeting more than one node we should expect as many responses out of the response 
// channels as there are targeted nodes
select {
case response := <-responseChan:
	fmt.Println("Success:", response.StandardOut)
case err := <-errChan:
	fmt.Println("Error:", err)
case <-time.After(300 * time.Millisecond):
	fmt.Println("This is taking too long!")
}
```

Alternatively you can manage your own [low-level communication](https://godoc.org/github.com/Jumpscale/agentcontroller2/newclient#LowLevelClient) by handling command construction and response parsing yourself.

```go
client := newclient.NewLowLevelClient("localhost:9999", "")

target := newclient.AllNodes()

// Command factories are here to help you construct various commands
command := commandfactory.CommandExecute(target, "ls", []string{"/opt"})

// Filter out all intermediate responses
responseChan := newclient.TerminalResponses(client.Execute(command))

// You'll be reciving intermediate and terminal responses from each targeted agent
for {
	// Recieve until channel is closed, or responses time-out
	
	select {
	case response, isOpen := <-responseChan:
	
		if !isOpen { 
			fmt.Println("No more responses to be received. We're done here.")
			return
		}
		
		fmt.Println("Got a response", &response)
		
	case <- time.After(500 * time.Millisecond):
		fmt.Println("I haven't received a response in so long! Are you sure everything is okay?")
	}
}
```
