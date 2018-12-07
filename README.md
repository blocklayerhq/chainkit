# ChainKit

[![CircleCI](https://circleci.com/gh/blocklayerhq/chainkit.svg?style=shield&circle-token=d1cf6680667cd473a3827610073c0678f280a207)](https://circleci.com/gh/blocklayerhq/chainkit)
[![GoDoc](https://godoc.org/github.com/blocklayerhq/chainkit?status.png)](https://godoc.org/github.com/blocklayerhq/chainkit)

ChainKit is a toolkit for blockchain development. It includes primitives for creating, building and running decentralized applications built on top of [Tendermint](https://tendermint.com/) and the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).

Key features:
- **Scaffold**: Generate all the [Tendermint](https://tendermint.com/) & [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) boilerplate automatically to get started in seconds.
- **Build and Run**: Under the hood, *chainkit* packages your app in a *Docker* container.
- **Testnet**: Anyone in the world can join your network by running one command. Under the hood, *chainkit* uses [IPFS](https://ipfs.io/) and [libp2p](https://libp2p.io/) to share data and discover peers.

## Installing

Requirements:
-   Go 1.11 or higher
-   A [working golang](https://golang.org/doc/code.html) environment

From this repository, run:
```bash
$ make
$ cp chainkit /usr/local/bin
```

## Usage

### Create, Build & Start

In order to create a new (empty) application, just run the following:
```bash
$ cd ~/go/src/github.com
$ chainkit create demoapp
```

You can then start by running:
```bash
$ cd demoapp
$ chainkit start
```

Then open [http://localhost:42001/](http://localhost:42001/) to see *Tendermint*'s RPC interface.

### Testnet

Anyone in the world can join your network by running:
```bash
$ chainkit join <network ID>
```

Under the hood, *chainkit* uses [IPFS](https://ipfs.io/) to transfer your network's manifest, genesis file and Docker image between nodes.

A built-in discovery mechanism (using [libp2p](https://libp2p.io/) DHT) allows nodes to discover themselves in a completely decentralized fashion.