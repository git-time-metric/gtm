[![Stories in Ready](https://badge.waffle.io/git-time-metric/gtm.png?label=ready&title=Ready)](https://waffle.io/git-time-metric/gtm)
[![Stories in Ready](https://badge.waffle.io/git-time-metric/gtm.png?label=ready&title=Ready)](https://waffle.io/git-time-metric/gtm)
# Git Time Metrics (GTM)
### Simple, seamless, lightweight time tracking for all your git projects

[![Build Status](https://travis-ci.org/git-time-metric/gtm.svg?branch=develop)](https://travis-ci.org/git-time-metric/gtm) [![Build status](https://ci.appveyor.com/api/projects/status/gj6tvm8njgwj0hqi?svg=true)](https://ci.appveyor.com/project/mschenk42/gtm)

Git Time Metrics (GTM) is a tool to automatically track time spent reading and working on code that you store in a Git repository. By installing GTM and using supported plug-ins for your favorite editors, you can immediately realize better insight into how you are spending your time and on what files.

GTM is automatic, seamless and lightweight.  There is no need to remember to start and stop timers.  It's a process that only runs on occasion to capture edit events triggered by the editor.  It does not require a background process or any file system monitoring.

Your time metrics are stored locally with the repository as [Git notes](https://git-scm.com/docs/git-notes). If you want to share your data,  it can be pushed and fetched to and from the remote repository. Other GTM users on your team can do the same. This provides you the ability to see time metrics for the entire team.

```
9361c18 Rename packages
Sun Jun 19 09:56:40 2016 -0500 Michael Schenk  34m 30s

341bd77 Vagrant file for testing on Linux
Sun Jun 19 09:43:47 2016 -0500 Michael Schenk  1h 16m  0s

792ba19 Require a 40 char SHA commit hash
Thu Jun 16 22:28:45 2016 -0500 Michael Schenk  1h  1m  0s
```
```
b2d16c8 Refactor discovering of paths when recording events
Thu Jun 16 11:08:47 2016 -0500 Michael Schenk

       30m 18s  [m] event/event.go
       12m 31s  [m] event/manager.go
        3m 14s  [m] project/project.go
        1m 12s  [r] .git/COMMIT_EDITMSG
        1m  0s  [r] .git/index
           25s  [r] event/manager_test.go
           20s  [r] metric/manager.go
       49m  0s
```
```
           0123456789012345678901234
Fri Jun 24 *                              22m  0s
Sat Jun 25 **                          1h 28m  0s
Sun Jun 26 ****                        3h 28m  0s
Mon Jun 27 *                               4m  0s
Tue Jun 28 **                          1h 36m  0s
                                       6h 58m  0s
```
# Getting Started

#### Install the latest GTM release

**Mac OS X***

The simplest way to install is to use [Homebrew](http://brew.sh).
```
brew tap git-time-metric/gtm
brew install gtm
```

**Linux or Mac OS X***

- Download the pre-build executable from [here](https://github.com/git-time-metric/gtm/releases/latest)
- Extract the tar file to /usr/local/bin, `tar -C /usr/local/bin -zxf <file.tar.gz>`
- Make sure it's executable, `chmod +x /usr/local/bin/gtm`
- Check to make sure `/usr/loca/bin` is in your PATH (it should be already there)
- Stay tuned, we will soon be distributing Linux packages


**Windows***

*Option 1*

- Download and run the windows installer from [here](https://github.com/git-time-metric/gtm/releases/latest)

*Option 2*

- Create a `gtm` directory in `c:\Program Files (x86)` and add it to your system's path
- Download the pre-build executable from [here](https://github.com/git-time-metric/gtm/releases/latest)
- Extract the tar file and install the `gtm.exe` in `c:\Program Files (x86)\gtm`
  - The release archive is in a tar format, here are some options documented on the Haskell site for [unpacking in Windows](https://wiki.haskell.org/How_to_unpack_a_tar_file_in_Windows)

**Note** - when running GTM in Windows, Mac OS X or Linux, if you recieve an error about something related to [libssh2](https://www.libssh2.org), please create an issue so we can look into it.

#### Install a GTM plug-in for your favorite editor

Currently we have plug-ins for Atom, Sublime and Vim. More will be added shortly and we are looking for others to contribute plug-ins.

- [Sublime 3](https://github.com/git-time-metric/gtm-sublime3-plugin)
- [Atom](https://github.com/git-time-metric/gtm-atom-plugin)
- [Vim](https://github.com/git-time-metric/gtm-vim-plugin)

#### Initialize a project for time tracking

```sh
cd ~/some-project
gtm init
```

#### Edit your files

Spend some time editing your files.  Check your progress with `gtm status`.

#### Commit your work

When you are ready, commit your work like you usually do.  GTM will automatically save the time spent associated with your commit. To check the time of the last commit type `gtm report`.

# Command Line Interface

GTM provides a simple CLI for initializing tracking and reporting on your time.

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
    Usage: gtm report [-n] [-format commits|totals|files|timeline] [-total-only] <commit-sha1...>
    Report on time logged

status
    Usage: gtm status [-total-only]
    Show time spent for working or staged files
```

For general help `gtm --help` and for help on a command `gtm --help <command>`

**Some examples on how to use the CLI**

*Report on the last 5 commits*

```
gtm report -n 5
```

*Report on the last 5 commits as a timeline*

```
gtm report -n 5 -format timeline
```

*Report on a specific commit SHA1*

```
gtm report 2401f73324c677f88fd40d2b434f2d007ce0b6f3 93f57ee594c917b2a372e06f09fa22009a145aac
```

*Pipe output from git log*

```
 git log --pretty=%H --since="8 days ago" | gtm report
```

**Pushing and fetching notes to and from the remote repository**

GTM adds a couple of [git aliases](https://git-scm.com/book/en/v2/Git-Basics-Git-Aliases) to make this easy.  It defaults to origin for the remote repository.

```
git fetchgtm
git pushgtm
```

Note - if you don't push your notes to the remote repository you will lose them if you delete and re-clone a repository.

# Known Issues

- Currently `git stash` is not fully supported.  You can certainly use git stash but your time may be assigned to the wrong commit. We will be adding full support for stashing in the near future.

# Contributing

GTM has reached beta status for the initial release but we are looking for others to help make it great. We also need to expand the editor plug-in library.

The plug-ins are very simple to write. Take a look at the [Atom](https://github.com/git-time-metric/gtm-atom-plugin), [Vim](https://github.com/git-time-metric/gtm-vim-plugin) and [Sublime 3](https://github.com/git-time-metric/gtm-sublime3-plugin) plug-ins to see how easy it is to create plug-ins.

For more detail on how to write plug-ins, check out the [Wiki](https://github.com/git-time-metric/gtm/wiki/Editor-Plug-ins).
