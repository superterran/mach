# Mach

[Website](https://superterran.net/mach) |
[Discussions](https://github.com/superterran/mach/discussions) |
[Documentation](https://github.com/superterran/mach/wiki) |
[Twitter](https://twitter.com/superterran) |
[Installation Guide](https://github.com/superterran/mach/wiki/Installation) |
[Contribution Guide](CONTRIBUTING.md)

Mach is a cli application for using docker to *quickly* and *easily* manage infrastructure and services through code.

[![GoDoc](https://godoc.org/github.com/gohugoio/hugo?status.svg)](https://pkg.go.dev/github.com/superterran/mach)
[![Go](https://github.com/superterran/mach/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/superterran/mach/actions/workflows/go.yml) 
[![Go Report Card](https://goreportcard.com/badge/github.com/superterran/mach)](https://goreportcard.com/report/github.com/superterran/mach)
[![codecov](https://codecov.io/gh/superterran/mach/branch/main/graph/badge.svg?token=S48U2MJP9I)](https://codecov.io/gh/superterran/mach)

Mach provides tooling around a simple docker and docker-machine based workflow for managing infrastructure, services and docker images. This allows you to easily leverage any git repository for the purposes of managing your Infrastructure as Code. 

This project is written in golang, using [cobra](https://github.com/spf13/cobra). Check the [wiki](https://github.com/superterran/mach/wiki) for additional documentation and user guides. 
  
# Usage

This tool runs in git repos thats meant to represent an IaC implementation. The git repo can store _docker images_, these are used to populate a registry and to as services for _docker compositions_. The repo can also store _docker compositions_ (or stacks), which can be used to deploy to _docker-machines_. Mach can also be used to transfer _docker-machine certificates_ to and from S3. 

```bash
mach build # builds every image in working directory (add .mach.yaml to configure)
mach build example # builds every image in `example` directory
mach build example:template # builds `Dockerfile-template[.tpl]` in `example` directory 
mach compose up # runs `docker-compose up` against every composition in working directory (add .mach.yaml to configure)
mach compose <service> up # runs `docker-compose up` against composition that matches the service
mach machine restore example-restore # downloads machine from S3 and installs to ~/.docker/machine
mach machine backup example-machine # copies machine configuration and certs to S3
```

## Building Docker Images

Maintain a collection of docker images that can be rapidly [built and pushed](https://github.com/superterran/mach/wiki/Build-Command) to a registry. Dockerfiles can be made using templates supporting includes, conditionals, loops, etc. `mach build` can build these images, and tag them based on git branch and filename conventions. This allows for maintaining a mainline image for public use, and versions for test. 

## Managing Docker Machines

Mach can be used to backup docker-machine certificates and configurations to Amazon S3 buckets. This makes sharing docker-machine credentials with teammates (and pipelines) simple.

AWS authentication is performed through the golang library, which provides a variety of ways to authenticate. You can use a tool like `aws-vault`, `~/.aws/credentials` files or environment variables such as:

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_DEFAULT_REGION=us-west-2
```
## Managing Docker Compositions

This tool also provides a thin wrapper around the docker-compose command, and will process docker-compose.yml.tpl files before passing them to compose. The compose command can run against any one composition, or against all of them in sequence to allow for managing everything in one command. 

# Installation 



## Brew (Linux/Mac)

Installing with [brew](https://brew.sh/) is a quick way to get started...

```/bin/bash
brew tap superterran/mach
brew install mach
```

## Github Releases

Binaries are compiled with every release, you can grab it from the [releases](https://github.com/superterran/mach/releases/) page, and use it as-is. 

You can copy it to any directory and run it directly, or commit it and invoke directly i.e. `bin/mach`. 

It can also be installed to a $PATH and used globally i.e. 

```bash
cp ~/Downloads/mach /usr/local/bin/mach && chmod +x /usr/local/bin/mach
```

## Compiling Manually

If you prefer to compile from source, the Makefile can be used:

```bash
git clone git@github.com:superterran/mach.git 
cd mach
make install # runs `go build .` and copies to /usr/local/bin
```

# Contributing

For a complete guide to contributing to Mach, see the [Contribution Guide](CONTRIBUTING.md)

Bug reports and pull requests are welcome on GitHub at https://github.com/superterran/mach/issues. 

# License
Mach is released under the [MIT License](LICENSE)

# Author Information
This project was started in 2021 by Doug Hatcher <superterran@gmail.com>.
