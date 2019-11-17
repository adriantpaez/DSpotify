## Overview

Decentralized Spotify based on Kademlia DHT

## Installation

### Compiling

```shell script
cd $GOPATH/src
git clone https://github.com/stdevAdrianPaez/DSpotify.git
cd DSpotify
make build
```

After that, you have on current folder a binary named `dspotify`

### Running

For run DSpotify node the binary accept some params:

 - `idSeed`: This param is required. The seed of NodeID, `NodeID = SHA1(seed)`
 -  `inPort`,`outPort`: Ports used by node for connect with other nodes and then build a P2P network

#### Example 1

```shell script
dspotify --idSeed HelloWorld
```

**Output**

```shell script
YY/MM/DD HH:MM:SS Starting DSpotify server
ID: db8ac1c259eb89d4a131b253bacfca5f319d54f2
IP: 127.0.0.1 InPort: 8000 OutPort: 8001
```

Observe the `ID: db8ac1c259eb89d4a131b253bacfca5f319d54f2` is the result of `SHA1(seed)`.
And `InPort`, `OutPort` have values `8000` and `8001` by default.

#### Example 2

```shell script
dspotify --idSeed HelloWorld --inPort 8080 --outPort 8081
```

**Output**
```shell script
YY/MM/DD HH:MM:SS Starting DSpotify server
ID: db8ac1c259eb89d4a131b253bacfca5f319d54f2
IP: 127.0.0.1 InPort: 8080 OutPort: 8081
```

On this example the `ID` is the same of Example 1 because seed is equal (`HelloWorld`). The only changes is `InPort`
and `OutPort`.

