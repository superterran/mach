# Mach

[Website](https://superterran.net/mach) |
[Discussions](https://github.com/superterran/mach/discussions) |
[Documentation](https://github.com/superterran/mach/wiki) |
[Twitter](https://twitter.com/superterran) |
[Installation Guide](https://github.com/superterran/mach/wiki/Installation) |
[Contribution Guide](CONTRIBUTING.md)

A cli application for using docker to *quickly* and *easily* manage infrastructure and services through code.

[![GoDoc](https://godoc.org/github.com/gohugoio/hugo?status.svg)](https://pkg.go.dev/github.com/superterran/mach)
[![Go](https://github.com/superterran/mach/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/superterran/mach/actions/workflows/go.yml) 
[![Go Report Card](https://goreportcard.com/badge/github.com/superterran/mach)](https://goreportcard.com/report/github.com/superterran/mach)
[![codecov](https://codecov.io/gh/superterran/mach/branch/main/graph/badge.svg?token=S48U2MJP9I)](https://codecov.io/gh/superterran/mach)


# Overview

Mach provides tooling around a simple docker and docker-machine based workflow for managing infrastructure, services and docker images. This allows you to easily leverage any git repository for the purposes of managing your Infrastructure as Code. 

This project is written in golang, using [cobra](https://github.com/spf13/cobra). Check the [wiki](https://github.com/superterran/mach/wiki) for additional documentation and user guides. 
  
## Installation

Mach is a simple command-line tool and only one executable. 

* You can grab it from the [releases](https://github.com/superterran/mach/releases/) page, and use it as-is. 
    * You can also copy it to any directory and run it directly. You can even commit it and invoked directly i.e. `bin/mach`
    * It can also be installed to a $PATH and used globally i.e. `cp ~/Downloads/mach /usr/local/bin/mach && chmod +x /usr/local/bin/mach`

If you prefer to compile from source, the Makefile can be used:

```bash
git clone git@github.com:superterran/mach.git 
cd mach
make install # runs `go build .` and copies to /usr/local/bin
```
## Managing Docker Machines

Mach can be used to backup docker-machine certificates and configurations to Amazon S3 buckets. This makes sharing docker-machine credentials with teammates (and pipelines) simple.

AWS authentication is performed through the golang library, which provides a variety of ways to authenticate. You can use a tool like `aws-vault`, `~/.aws/credentials` files or environment variables such as:

```
$ export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
$ export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
$ export AWS_DEFAULT_REGION=us-west-2
```
## Building Docker Images

Maintain a collection of docker images that can be rapidly [built and pushed](https://github.com/superterran/mach/wiki/Build-Command) to a registry. Dockerfiles can be made using templates supporting includes, conditionals, loops, etc. `mach build` can build these images, and tag them based on git branch and filename conventions. This allows for maintaining a mainline image for public use, and versions for test. 

# Contributing

For a complete guide to contributing to Mach, see the [Contribution Guide](CONTRIBUTING.md)

Bug reports and pull requests are welcome on GitHub at https://github.com/superterran/mach/issues. 

# License
Mach is released under the [MIT License](LICENSE)

# Author Information
This project was started in 2021 by Doug Hatcher <superterran@gmail.com>.
