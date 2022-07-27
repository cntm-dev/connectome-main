
<h1 align="center">Ontology </h1>
<p align="center" class="version">Version 0.7.0 </p>

[![GoDoc](https://godoc.org/github.com/cntmio/cntmology?status.svg)](https://godoc.org/github.com/cntmio/cntmology)
[![Go Report Card](https://goreportcard.com/badge/github.com/cntmio/cntmology)](https://goreportcard.com/report/github.com/cntmio/cntmology)
[![Travis](https://travis-ci.org/cntmio/cntmology.svg?branch=master)](https://travis-ci.org/cntmio/cntmology)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/cntmio/cntmology?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)


English | [中文](install_CN.md) 
## Build development environment
The requirements to build Ontology are:

- Golang version 1.9 or later
- Glide (a third party package management tool)
- Properly configured Go language environment
- Golang supported operating system

## Deployment|Get Ontology
### Get from source code

Clone the Ontology repository into the appropriate $GOPATH/src/github.com/cntmio directory.

```
$ git clone --recursive https://github.com/cntmio/cntmology.git
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
$ make
```

After building the source code sucessfully, you should see two executable programs:

- `cntmology`: the node program/command line program for node ccntmrol

### get from release
You can download at [release page](https://github.com/cntmio/cntmology/releases).