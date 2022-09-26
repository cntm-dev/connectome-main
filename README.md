
<h1 align="center">Ontology </h1>
<h4 align="center">Version 1.0 </h4>

[![GoDoc](https://godoc.org/github.com/cntmio/cntmology?status.svg)](https://godoc.org/github.com/cntmio/cntmology)
[![Go Report Card](https://goreportcard.com/badge/github.com/cntmio/cntmology)](https://goreportcard.com/report/github.com/cntmio/cntmology)
[![Travis](https://travis-ci.org/cntmio/cntmology.svg?branch=master)](https://travis-ci.org/cntmio/cntmology)
[![Discord](https://img.shields.io/discord/102860784329052160.svg)](https://discord.gg/gDkuCAq)

English | [中文](README_CN.md)

Welcome to Ontology's source code repository!

Ontology is dedicated to creating a modularized, freely configurable, interoperable cross-chain, high-performance, and horizcntmally scalable blockchain infrastructure system. Ontology makes deploying and invoking decentralized applications easier.

The code is currently alpha quality, but it is in the process of rapid development. The master code may be unstable; stable versions can be downloaded in the release page.

The public test network is described below. We sincerely welcome and hope more developers join Ontology.

## Features

- Scalable lightweight universal smart ccntmract
- Scalable WASM ccntmract support
- Crosschain interactive protocol (processing)
- Multiple encryption algorithm support
- Highly optimized transaction processing speed
- P2P link layer encryption (optional module)
- Multiple consensus algorithm support (VBFT/DBFT/RBFT/SBFT/PoW)
- Quick block generation time


## Ccntments

- [Build development environment](#build-development-environment)
- [Get Ontology](#get-cntmology)
    - [Get from source code](#get-from-source-code)
    - [Get from release](#get-from-release)
- [Server deployment](#server-deployment)
    - [Select network](#select-network)
        - [Mainnet sync node deployment](#mainnet-sync-node-deployment)
        - [Public test network Polaris sync node deployment](#public-test-network-polaris-sync-node-deployment)
        - [Single-host deployment configuration](#single-host-deployment-configuration)
        - [Multi-hosts deployment configuration](#multi-hosts-deployment-configuration)
            - [VBFT Deployment](#vbft-deployment)
            - [DBFT Deployment](#dbft-deployment)
        - [Deploy Completed](#deploy-completed)
    - [Implement](#implement)
        - [Run in docker](#run-in-docker)
    - [cntm transfer sample](#cntm-transfer-sample)
    - [Query transfer status sample](#query-transfer-status-sample)
    - [Query account balance sample](#query-account-balance-sample)
- [Ccntmributions](#ccntmributions)
- [Open source community](#open-source-community)
    - [Site](#site)
    - [Developer Discord Group](#developer-discord-group)
- [License](#license)

## Build development environment
The requirements to build Ontology are:

- Golang version 1.9 or later
- Glide (a third party package management tool)
- Properly configured Go language environment
- Golang supported operating system

## Get Ontology
### Get from source code

Clone the Ontology repository into the appropriate $GOPATH/src/github.com/cntmio directory.

```
$ git clone https://github.com/cntmio/cntmology.git
```
or
```
$ go get github.com/cntmio/cntmology
```
Fetch the dependent third party packages with glide.

```
$ cd $GOPATH/src/github.com/cntmio/cntmology
$ glide install
```

Build the source code with make.

```
$ make all
```

After building the source code sucessfully, you should see two executable programs:

- `cntmology`: the node program/command line program for node ccntmrol
- `tools/sigsvr`: (optional)Ontology Signature Server - sigsvr is a rpc server for signing transactions for some special requirement.detail docs can be reference at [link](./docs/specifications/sigsvr.md)

### Get from release
- You can download latest cntmology binary file with ` curl https://dev.cntm.io/cntmology_install | sh `.

- You can download other version at [release page](https://github.com/cntmio/cntmology/releases).

## Server deployment
### Select network
To run Ontology successfully,  nodes can be deployed by four ways:

- Mainnet sync node deployment
- Public test network Polaris sync node deployment
- Single-host deployment
- Multi-hosts deployment

#### Mainnet sync node deployment

Run cntmology straightly

   ```
	./cntmology --networkid 1
   ```


#### Public test network Polaris sync node deployment

Run cntmology straightly

   ```
	./cntmology --networkid 2
   ```


#### Single-host deployment configuration

Create a directory on the host and store the following files in the directory:
- Node program + Node ccntmrol program  `cntmology`
- Wallet file`wallet.dat`

Run command `$ ./cntmology --testmode --networkid 3` can start single-host test net.

Here's a example of single-host configuration:

- Directory structure

    ```shell
    $ tree
    └── cntmology
        ├── cntmology
        └── wallet.dat
    ```

#### Multi-hosts deployment configuration

Note: When you want to build a private net to run cntmology in DBFT or VBFT, you must use --config argument to specify a configuration file, and
use --networkid to define a net work identity of your network(not equals 1/2/3), otherwise the node will link to mainnet by default.

##### VBFT Deployment

In the multi-hosts enviroment, we need 7 nodes to run cntmology at least in VBFT.

We can perform a quick deployment by modifying the configuration file [`config-vbft.json`](./docs/specifications/config-vbft.json), 
click [here](./docs/specifications/config.md) to read instruction of config file.

1. Generate 7 wallet file, each wallet ccntmains an account. These account is bookkeepers of consensus. The account generated by :
	```
	./cntmology account add -d -w wallet.dat
	Use default setting '-t ecdsa -b 256 -s SHA256withECDSA' 
		signature algorithm: ecdsa 
		curve: P-256 
		signature scheme: SHA256withECDSA 
	Password:
	Re-enter Password:

	Index: 1
	Label: 
	Address: AXkDGfr9thEqWmCKpTtQYaazJRwQzH48eC
	Public key: 03d7d8c0c4ca2d2bc88209db018dc0c6db28380d8674aff86011b2a6ca32b512f9
	Signature scheme: SHA256withECDSA

	Create account successfully.
	```
    use -w argument to define wallet file name.

2. Modify `config-vbft.json`, set public key and address of 7 accounts generated in last step into peers config in `config-vbft.json`.

3. Copy related file into target host, including:

   - A configuration file`config-vbft.json`
   - Node program`cntmology`
   - wallet file
   
4. Set the network connection port number for each node (recommend using the default port configuration, instead of modifying)

   - `NodePort`is P2P connection port number (default: 20338)
   - `HttpJsonPort` and `HttpLocalPort` are RPC port numbers (default: 20336, 20337)

5. Seed nodes configuration

   - Select at least one seed node out of 7 hosts and fill the seed node address into the `SeelList` of each configuration file. The format is `Seed node IP address + Seed node NodePort`.

##### DBFT Deployment

In the multi-hosts enviroment, we need 4 nodes to run cntmology at least in DBFT.

We can perform a quick deployment by modifying the configuration file [`config-dbft.json`](./docs/specifications/config-dbft.json),
 click [here](./docs/specifications/config.md) to read 
 instruction of config file.
1. Copy related file into target host, including:
  
     - Configuration file`config-dbft.json`
     - Node program`cntmology`
     
2. Set the network connection port number for each node (recommend using the default port configuration, instead of modifying)

   - `NodePort`is P2P connection port number (default: 20338)
   - `HttpJsonPort` and `HttpLocalPort` are RPC port numbers (default: 20336, 20337)

3. Seed nodes configuration
  
      - Select at least one seed node out of 4 hosts and fill the seed node address into the `SeelList` of each configuration file. The format is `Seed node IP address + Seed node NodePort`.

4. Create wallet file

   - Through command line program, on each host create wallet wallet.dat needed for node implementation.
        ```
        ./cntmology account add -d -w wallet.dat
        Use default setting '-t ecdsa -b 256 -s SHA256withECDSA' 
        signature algorithm: ecdsa 
        curve: P-256 
        signature scheme: SHA256withECDSA 
        Password:
        Re-enter Password:
            
        Index: 1
        Label: 
        Address: AXkDGfr9thEqWmCKpTtQYaazJRwQzH48eC
        Public key: 03d7d8c0c4ca2d2bc88209db018dc0c6db28380d8674aff86011b2a6ca32b512f9
        Signature scheme: SHA256withECDSA
            
        Create account successfully.
        ```

5. Bookkeepers configuration

   - While creating a wallet for each node, the public key information of the wallet will be displayed. Fill in the public key information of all nodes in the `Bookkeepers` field of each node's configuration file.

     Note: The public key information for each node's wallet can also be viewed via the command line program:

        ```
        1	AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN (default)
        	Label: 
        	Signature algorithm: ECDSA
        	Curve: P-256
        	Key length: 384 bits
        	Public key: 030e5d50bf585ff5c73464114244b93f04b231862d6bbdfd846be890093b2c1c17
        	Signature scheme: SHA256withECDSA
        ```

#### Deploy Completed

Now multi-host configuration is completed, directory structure of each node is as follows:

   ```shell
	$ ls
	config.json cntmology wallet.dat
   ```

### Implement

Run each node program in any order and enter the node's wallet password after the `Password:` prompt appears.

If you wish to run a consensus node (such as in a private net), the --enableconsensus argument must be used. If you want to run
a private net, use --networkid argument to specify your net work identify（not equals 1/2/3） and use --config argument to specify your configuration
file.

such as:
   ```
    $ ./cntmology --enableconsensus --networkid 4 --config ./config.json
    $ - Input your wallet password
   ```

Run `./cntmology --help` for details, also you can read [cntmology CLI user guide](./docs/specifications/cli_user_guide.md) to get more information.


#### Run in docker

Please ensure there are docker environment in your machine.

1. make docker image

    - In the root directory of source code，run`make docker`, it will make cntmology image in docker.

2. run cntmology image

    - Use command `docker run cntmio/cntmology`to run cntmology；

    - If you need to allow interactive keyboard input while the image is running, you can use the `docker run -ti cntmio/cntmology` command to start the image;

    - If you need to keep the data generated by image at runtime, you can refer to the data persistence function of docker (e.g. valume);

    - If you need to add cntmology parameters, you can add them directly after `docker run cntmio/cntmology` such as `docker run cntmio/cntmology --networkid 2`.
     The parameters of cntmology command line refer to [here](./docs/specifications/cli_user_guide.md).

### cntm transfer sample
 -- from: transfer from； -- to: transfer to； -- amount: cntm amount；
```shell
  ./cntmology asset transfer  --to=AXkDGfr9thEqWmCKpTtQYaazJRwQzH48eC --amount=10
```
If transfer asset successd, the result will show as follow:

```
Transfer cntm
From:TA6edvwgNy3c1nBHgmFj8KrgQ1JCJNhM3o
To:TA4Xe9j8VbU4m3T1zEa1uRiMTauiAT88op
Amount:10
TxHash:10dede8b57ce0b272b4d51ab282aaf0988a4005e980d25bd49685005cc76ba7f
```
TxHash is the transfer transaction hash, we can query transfer result by txhash.
Because of generate block time, the transfer transaction will not execute befer at least generate one block.

### Query transfer status sample

--hash:transfer transaction hash
```shell
./cntmology asset status --hash=10dede8b57ce0b272b4d51ab282aaf0988a4005e980d25bd49685005cc76ba7f
```
result：
```shell
Transaction:transfer success
From:AXkDGfr9thEqWmCKpTtQYaazJRwQzH48eC
To:AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN
Amount:10
```

### Query account balance sample

--address: account address

```shell
./cntmology asset balance --address=AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN
```
result：
```shell
BalanceOf:AYiToLDT2yZuNs3PZieXcdTpyC5VWQmfaN
cntm:10
cntm:0
cntmApprove:0
```

## Ccntmributions

Please open a pull request with a signed commit. We appreciate your help! You can also send your code as emails to the developer mailing list. You're welcome to join the Ontology mailing list or developer forum.

Please provide detailed submission information when you want to ccntmribute code for this project. The format is as follows:

Header line: explain the commit in one line (use the imperative).

Body of commit message is a few lines of text, explaining things  in more detail, possibly giving some background about the issue  being fixed, etc.

The body of the commit message can be several paragraphs. Please do proper word-wrap and keep columns shorter than 74 characters or so. That way "git log" will show things  nicely even when it is indented.

Make sure you explain your solution and why you are doing what you are  doing, as opposed to describing what you are doing. Reviewers and your  future self can read the patch, but might not understand why a  particular solution was implemented.

Reported-by: whoever-reported-it &
Signed-off-by: Your Name [youremail@yourhost.com](mailto:youremail@yourhost.com)

## Open source community
### Site

- <https://cntm.io/>

### Developer Discord Group

- <https://discord.gg/4TQujHj/>

## License

The Ontology library is licensed under the GNU Lesser General Public License v3.0, read the LICENSE file in the root directory of the project for details.
