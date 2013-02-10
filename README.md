# ezconnector

This code includes a trivial server and client that will allow clients to create UDP connections to each other even if they are protected or hidden by a NAT router. The server merely serves, effectively, as a STUN server to allow clients to negotiate a way of connecting to each other. After that succeeds, everything else is peer-to-peer.

This is a naive implementation with little-to-no error handling, especially if the tunneling works. You may wish to fallback to TURN or some other connection mechanism.

This was originally created to debug an issue I was having with a dependency (David Anderson's nat.go package). I had missed a particular regarding copying slices, versus copying into them. This code lives on in a more complicated fashion in my goxpn [1] project.

[1] goxpn isn't publicly visible yet and when it does become publicly available, it may exist under a different name. Hopefully I remember to update this accordingly.

## Example Usage

1. go get github.com/colemickens/ezconnector
2. (in one terminal) ezconnector --server=localhost:9000
3. (in another) ezconnector --client=localhost:9000
4. (on another computer) ezconnector --cleint=192.168.1.118:9000

You will see the second client "call" the first client. They will exchange packets until they find a route to each other via David Andersen's nat traversal code.