# Mach

[Website](https://superterran.net/mach) |
[Discussions](https://github.com/superterran/mach/discussions) |
[Documentation](https://github.com/superterran/mach/wiki) |
[Twitter](https://twitter.com/superterran) |
[Installation Guide](https://github.com/superterran/mach/wiki/Installation) |
[Contribution Guide](CONTRIBUTING.md)

A cli application for using docker to *quickly* and *easily* manage infrastructure and services.

[![Go](https://github.com/superterran/mach/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/superterran/mach/actions/workflows/go.yml)

# Overview

Mach provides some tooling around a simple docker and docker-machine based workflow for managing infrastructure, services and docker images. This allows you to easily leverage any git repository for the purposes of managing your Infrastructure as Code. 

## Installation

Mach is a simple command-line tool and only one executable. 

* You can grab it from the [releases](https://github.com/superterran/mach/releases/) page, and use it as-is. 
    * It can be copied to a working directory, even commited, and invoked directly i.e. `bin/mach`
    * It can also be installed to a PATH and used globally i.e. `cp ~/Downloads/mach /usr/local/bin/mach && chmod +x /usr/local/bin/mach`. 

You can also use the bundled Makefile to install it manually: 

```bash
git clone git@github.com:superterran/mach.git 
cd mach
make install # runs `go build .` and copies to /usr/local/bin
```
## Building Docker Images

Maintain a collection of docker images that can rapidly [built and pushed](https://github.com/superterran/mach/wiki/Build-Command) to a registry. Dockerfiles can be made using templates supporting includes, conditionals, loops, etc. `mach build` can build these images, and tag them based on git branch and filename conventions. This allows for maintaining a mainline image for public use, and versions for test. 

written in golang, using [cobra](https://github.com/spf13/cobra). Check the [wiki](https://github.com/superterran/mach/wiki) for additional documentation. 
  

# Contributing

For a complete guide to contributing to Mach, see the [Contribution Guide](CONTRIBUTING.md)

Bug reports and pull requests are welcome on GitHub at https://github.com/superterran/mach/issues. 

# License
Mach is released under the [MIT License](LICENSE)

# Author Information
This project was started in 2021 by Doug Hatcher <superterran@gmail.com>.