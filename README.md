# Introduction

`fptrace` is a Linux process tracing tool that records process launches and file accesses.  Results can be saved in a `deps.json` file or used to generate launcher scripts.  It works like `strace` but produces machine readable output and resolves relative pathnames into absolute ones.  Optionally it also records environment variables and prevents deletions.

# `deps.json`

`fptrace -d deps.json sh -c 'echo a > a; cat a | tee b'` in `/tmp` makes:

```json
[
  {
    "Cmd": {
      "Parent": 1, "ID": 2,
      "Dir": "/tmp", "Path": "/bin/cat", "Args": ["cat", "a"]
    },
    "Inputs": ["/etc/ld.so.cache", "/lib/x86_64-linux-gnu/libc.so.6", "/tmp/a"],
    "Outputs": []
  },
  {
    "Cmd": {
      "Parent": 1, "ID": 3,
      "Dir": "/tmp", "Path": "/usr/bin/tee", "Args": ["tee", "b"],
    },
    "Inputs": ["/etc/ld.so.cache", "/lib/x86_64-linux-gnu/libc.so.6"],
    "Outputs": ["/tmp/b"]
  },
  {
    "Cmd": {
      "Parent": 0, "ID": 1,
      "Dir": "/tmp", "Path": "/bin/sh", "Args": ["sh", "-c", "echo a > a; cat a | tee b"],
    },
    "Inputs": ["/etc/ld.so.cache", "/lib/x86_64-linux-gnu/libc.so.6"],
    "Outputs": ["/tmp/a"]
  }
]
```

The result is a list of command executions (ordered by the time of their exit): an execution begins with an `execve` and ends with the last spawned thread or fork.

- `ID` is a unique execution identifier (counting from 1)
- `Parent` is the `ID` of the execution that spawned it
- `Dir` is the initial working directory
- `Path` is an absolute path to the executable
- `Args` are `execve` arguments

`Inputs` and `Outputs` list absolute paths to files opened for reading and writing, except that files opened for writing and later opened for reading are not listed as execution `Inputs`.

# Launcher scripts

`fptrace -s /tmp/scripts sh -c 'echo a > a; cat a | tee b'` generates `0-1-sh`, `1-2-cat`, and `1-3-tee`:

- `0-1-sh`
```sh
#!/bin/sh
cd /tmp
exec sh -c 'echo a > a; cat a | tee b'
```
- `1-2-cat`
```sh
#!/bin/sh
cd /tmp
exec cat a
```
- `1-3-tee`
```sh
#!/bin/sh
cd /tmp
exec tee b
```

# Installation

With go get:
```sh
go get github.com/orivej/fptrace
go generate github.com/orivej/fptrace
```

On NixOS:
```sh
nix-env -if https://github.com/orivej/fptrace/archive/master.tar.gz
```
