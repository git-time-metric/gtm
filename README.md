## Overview

![Git Time Metrics](http://dogsbuttbrew.com/gtm/img/GTMLogoShort.png)

Git Time Metrics (gtm) is a tool to automatically track time spent reading and
working on code that you store in a GIT repository. By installing GTM and
using supported plugins to your favorite editors, you can immediately realize
better insight into how you are spending your time and on what files.

## Installation

### Installation from a Package Manager

GTM is available from several package managers, including:

* [Homebrew for Mac](http://brew.sh/)
* [Advanced Packaing Tool for Debian and offshoots](https://wiki.debian.org/Apt)

### Building from Source

If GTM is not available from a package manager for your platform, you can always
build it from source! GO's build tool chain makes it easy to compile and build
on any supported environment.

To build GTM, follow these steps:

1. Create an _edgegio_ directory in your $GOPATH/src directory

    ```
$ cd $GOPATH/src
$ mkdir edgegio
    ```

1. Clone the repository

    ```
$ git clone https://gradymke@bitbucket.org/edgegio/gtm.git
    ```

1. Build and install the executable

    ```
$ cd edgegio
$ go build
$ go install
    ```

## Initializing your GIT Repo

Once installed, you need to initialize any GIT repositories that you wish to
track with GTM. It's easy to do so. Just go to the base of the project you wish
to track and issue the following command:

```
$ gtm init
```

## Find and Install Plugins

GTM gathers its data through editor plugins. To find a plugin, go to the [Plugin Repository](http://dogsbuttbrew.com/gtm/plugins/) and look for one for your favorite editor. If your editor has a plugin or extension repository of its own, you can check there as well. Just search for _GTM_.

## Reporting Bugs

Report bugs by logging an issue to this repository's issue tracker.

## Getting help

## More Information

You can find more information on GTM and how you can contribute to the project,
either through pull requests or by developing a plug-in by navigating to the
[Git Time Metrics website](http://dogsbuttbrew.com/gtm/).
