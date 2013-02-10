# ezconnector

This code includes a trivial server and client that will allow clients to create UDP connections to each other even if they are protected or hidden by a NAT router. The server merely serves as a middle man and directory service. It allows users to connect, find out about other users, and then establish peer-to-peer connections that don't go through the server.

This is a naive implementation with little-to-no error handling, especially if the tunneling works. You may wish to fallback to TURN or some other connection mechanism.

This was originally created to debug an issue I was having with a dependency (David Anderson's nat.go package). I had missed a particular regarding copying slices, versus copying into them.

## Example Usage

1. `go get github.com/colemickens/ezconnector`
2. (in one terminal) `ezconnector --server=localhost:9000`
3. (in another) `ezconnector --client=localhost:9000`
4. (on another computer) `ezconnector --client=192.168.1.118:9000`

You will see the second client "call" the first client. They will exchange packets until they find a route to each other via David Anderson's nat traversal code.