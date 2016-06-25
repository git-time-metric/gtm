# Git Time Metrics (GTM)
### Simple, seamless, lightweight time tracking for all your git projects

[![Build Status](https://travis-ci.org/git-time-metric/gtm.svg?branch=develop)](https://travis-ci.org/git-time-metric/gtm) [![Build status](https://ci.appveyor.com/api/projects/status/gj6tvm8njgwj0hqi?svg=true)](https://ci.appveyor.com/project/mschenk42/gtm)

Git Time Metrics (GTM) is a tool to automatically track time spent reading and working on code that you store
in a Git repository. By installing GTM and using supported plugins to your favorite editors, you can immediately
realize better insight into how you are spending your time and on what files.

GTM has reached beta status for the initial release but we are looking for others to help make it great. We
also need to expand the editor plugin library.

The plugins are very simple to write. Take a look at the Atom, Vim and Sublime plugins to see how easy it is to
create plugins.

## Initializing your Git project for time tracking

Once installed, you need to initialize any Git repositories that you wish to track with GTM. It's easy to do so.
Just go to the base of the project you wish to track and issue the following command:
```
gtm init
```

Here's a list of all the commands available for GTM.
```
usage: gtm [--version] [--help] <command> [<args>]

Available commands are:
commit
    Usage: gtm commit [-yes]
    Save your logged time with the last commit
    This is automatically called from the postcommit hook
    Warning - any time logged will be cleared from your working directory

init
    Usage: gtm init
    Initialize a git project for time tracking

record
    Usage: gtm record <filepath>
    Record a file event

report
    Usage: gtm report [-n] [-format commits|totals|files|timeline] [-total-only]
    Report on time logged

status
    Usage: gtm status [-total-only]
    Show time spent for working or staged files
```

## Reporting
Here are some samples reports from GTM.

```
6c5a028 Standardize handling of file paths across operating systems
Tue Jun 14 19:53:55 2016 -0500 Michael Schenk

       36m 35s  [m] scm/git.go
       33m 54s  [m] util/test.go
       24m 34s  [m] metric/metric.go
       13m 23s  [m] metric/manager_test.go
       10m 45s  [m] event/manager_test.go
       10m  0s  [m] metric/metric_test.go
        9m 49s  [r] metric/manager.go
        5m  0s  [r] note/note_test.go
        5m  0s  [m] note/note.go
        3m 40s  [r] event/event.go
        3m  0s  [d] x/z.go
        1m  0s  [r] command/status.go
           20s  [r] project/project.go
    2h 37m  0s
```
```
           0123456789012345678901234
Fri Jun 17 **                          1h  3m  0s
Sat Jun 18 *                              12m  0s
Sun Jun 19 **                          1h 18m  0s
                                       2h 33m  0s
```
```
601d24c Conditionally install go and compile git2go and libgit2
Sun Jun 19 12:27:14 2016 -0500 Michael Schenk  42m 30s

9361c18 Rename packages
Sun Jun 19 09:56:40 2016 -0500 Michael Schenk  34m 30s

341bd77 Vagrant file for testing on Linux
Sun Jun 19 09:43:47 2016 -0500 Michael Schenk  1h 16m  0s

792ba19 Require a 40 char SHA commit hash
Thu Jun 16 22:28:45 2016 -0500 Michael Schenk  1h  1m  0s
```

## Installing From Binaries

Pre-built binaries can be download [here](https://github.com/git-time-metric/gtm/releases). The binary requires that [libssh2](https://www.libssh2.org) be installed on your system.
For OS X, you can `brew install libssh2`. Most systems already have this, so first try running GTM before installing this library.

## Installing From Source

GTM statically links with the [libgit2](https://libgit2.github.com/) library.  To compile GTM from source requires compiling the C library.
The easiest way to get up and running for development is to utilize Vagrant with the provided Vagrantfile. Otherwise follow the directions
for OS X and Linux to compile and install locally.

### OS X and Linux

```
go get -d github.com/libgit2/git2go
cd $GOPATH/src/github.com/libgit2/git2go
git checkout next
git submodule update --init
make install

go get -u github.com/git-time-metric/gtm
cd $GOPATH/src/github.com/git-time-metric/gtm
go get -t -v ./...
go test ./...
go install
```
