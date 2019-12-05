## Overview

## Index

- [Installation](#installation)
    - [Compiling](#compiling)
    - [Running](#running)   

Decentralized Spotify based on Kademlia DHT

## Installation

### Compiling

```
cd $GOPATH/src
git clone https://github.com/stdevAdrianPaez/DSpotify.git
cd DSpotify
make build
```

After that, you have on current folder a binary named `dspotify`

### Running

For run DSpotify node the binary accept some params:

 - `-database=string`: Path for database. (default `database.db`)
 - `-httpPort=int`: Port for listen requests to web API. (default `8080`)
 - `-idSeed=string`: [REQUIRED] Seed for NodeID. `NodeID = SHA1(idSeed)`
 - `-inPort=int`: Input port for listen all RPC requests. (default `8000`)
 - `-outPort=int`: Output port for send all RPC requests. (default `8001`)
 - `-ip=string`: The IP of the node. (default `127.0.0.1`)
 - `-known=string`: File with known contact for join to network

#### Example 1

```
dspotify -idSeed=HelloWorld
```

**Output**

```
YY/MM/DD HH:MM:SS Starting DSpotify server
ID: db8ac1c259eb89d4a131b253bacfca5f319d54f2
IP: 127.0.0.1 InPort: 8000 OutPort: 8001
WARNING: Not known contact
```

Observe the `ID: db8ac1c259eb89d4a131b253bacfca5f319d54f2` is the result of `SHA1(seed)`.
And `InPort`, `OutPort` have values `8000` and `8001` by default.

#### Example 2

```
dspotify -idSeed=HelloWorld -inPort=8080 -outPort=8081
```

**Output**
```
YY/MM/DD HH:MM:SS Starting DSpotify server
ID: db8ac1c259eb89d4a131b253bacfca5f319d54f2
IP: 127.0.0.1 InPort: 8080 OutPort: 8081
WARNING: Not known contact
```

On this example the `ID` is the same of Example 1 because seed is equal (`HelloWorld`). The only changes is `InPort`
and `OutPort`.

## API

Kademlia nodes implement a simple API to known stats of node. API server run on same host of node and run in port
`httpPort` (default `8080`)

### Endpoints

- `/buckets`: show complete list of buckets (finger table) state
- `/postman`: show state of node postman, opened connections with other nodes.
- `/contact`: contact of node, to use for other nodes to join to network

