`depgrapher` is a ptrace-based dependency grapher.  It runs a program under ptrace and produces `deps.json` like the following result for `sh -c 'echo a > a; cat a | tee b'`:

```json
[
  {
    "Cmd": {
      "Path": "/bin/cat", "Args": ["cat", "a"], "Dir": "/tmp",
      "ID": 2, "Parent": 1
    },
    "Inputs": ["/etc/ld.so.cache", "/lib/x86_64-linux-gnu/libc.so.6", "/tmp/a"],
    "Outputs": []
  },
  {
    "Cmd": {
      "Path": "/usr/bin/tee", "Args": ["tee", "b"], "Dir": "/tmp",
      "ID": 3, "Parent": 1
    },
    "Inputs": ["/etc/ld.so.cache", "/lib/x86_64-linux-gnu/libc.so.6"],
    "Outputs": ["/tmp/b"]
  },
  {
    "Cmd": {
      "Path": "/bin/sh", "Args": ["sh", "-c", "echo a > a; cat a | tee b"], "Dir": "/tmp",
      "ID": 1, "Parent": 0
    },
    "Inputs": ["/etc/ld.so.cache", "/lib/x86_64-linux-gnu/libc.so.6"],
    "Outputs": ["/tmp/a"]
  }
]
```

The result is a list of command executions.  (An execution begins with an `execve` and ends with the last spawned thread or fork.)  `Inputs` and `Outputs` list absolute paths to files opened for reading and writing.  (Except that files opened for writing and later opened for reading are not listed as `Inputs`.)  Command `Path` is an absolute path to the executable, `Args` are `execve` arguments, `Dir` is the initial working directory, `ID` is a unique execution identifier (counting from 1), and `Parent` is the `ID` of the execution that spawned it.
