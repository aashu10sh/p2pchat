# P2P Chat
A discord/slack like peer to peer chat application made to chat within a subnet with support for automatic host discovery, file transfer and video calling.

* Simple: well-defined, elegantly designed user facing application 
* Browser-Based: uses a browser window as user interface
* Isolated: chats across different wifi-networks are isolated

## Objective
Designed to be a effective chatting/calling alternative within the people in an office.

## Getting Started
The best way of getting started is to clone the repo and build the project locally, its dead simple.

### Getting p2pchat

```
git clone git@github.com:aashu10sh/p2pchat
```

Once the project is cloned

### Building the frontend

```bash
cd frontend
npm run build
```
Note: Any nodejs version above 22 should work just fine

The built frontend assets is embed inside the go binary and exposed via a file router.

### Building the backend
```bash
go build .
```


### Run application

```bash
./p2pchat
```

# Architecture

p2pchat's frontend is built in svelte with sv router.

p2pchat uses:
* go's net/http package for client/server communication between the backend and the frontend.
* grpc for node to node communication in a network, for sending texts, exchanging ICE candidates allowing webRTC communication to happen for video calls.
* mDNS for automatic discovery when a node comes to and leaves the network.
