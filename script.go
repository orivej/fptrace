package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	sh "github.com/djmitche/shquote"
	"github.com/orivej/e"
)

func writeScript(dir string, cmd Cmd) {
	name := fmt.Sprintf("%d-%d-%s", cmd.Parent, cmd.ID, path.Base(cmd.Path))
	f, err := os.OpenFile(path.Join(dir, name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777) //#nosec
	e.Exit(err)
	defer e.CloseOrPrint(f)

	interp, exec, cmdline := "#!/bin/sh", "exec", cmd.Args
	if cmd.Args[0] != cmd.Path {
		interp = "#!/usr/bin/env bash"
		exec = "exec -a " + sh.Quote(cmd.Args[0])
		cmdline = append([]string{cmd.Path}, cmd.Args[1:]...)
	}
	buf := bufio.NewWriter(f)
	fmt.Fprintln(buf, interp)
	fmt.Fprintf(buf, "\ncd %s\n", sh.Quote(cmd.Dir))
	if len(cmd.Env) != 0 {
		writeEnv(buf, cmd.Env)
	}
	fmt.Fprintf(buf, "\n${exec:-%s} %s \"$@\"\n", exec, sh.QuoteList(cmdline))
	err = buf.Flush()
	e.Exit(err)
}

func writeEnv(buf io.Writer, env []string) {
	env = append([]string{}, env...)
	sort.Strings(env)
	fmt.Fprintln(buf)
	for _, entry := range env {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) < 2 {
			fmt.Fprintf(buf, ": %s\n", sh.Quote(entry))
		}
		k, v := parts[0], parts[1]
		fmt.Fprintf(buf, "export %s=%s\n", sh.Quote(k), sh.Quote(v))
	}
}
