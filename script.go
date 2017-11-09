package main

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/djmitche/shquote"
	"github.com/orivej/e"
)

func writeScript(dir string, cmd Cmd) {
	name := fmt.Sprintf("%d-%d-%s", cmd.Parent, cmd.ID, path.Base(cmd.Path))
	f, err := os.OpenFile(path.Join(dir, name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777) //#nosec
	e.Exit(err)
	defer e.CloseOrPrint(f)

	sh, exec, cmdline := "#!/bin/sh", "exec", cmd.Args
	if cmd.Args[0] != cmd.Path {
		sh = "#!/usr/bin/env bash"
		exec = "exec -a " + shquote.Quote(cmd.Args[0])
		cmdline = append([]string{cmd.Path}, cmd.Args[1:]...)
	}
	buf := bufio.NewWriter(f)
	fmt.Fprintln(buf, sh)
	fmt.Fprintf(buf, "\ncd %s\n", shquote.Quote(cmd.Dir))
	if len(cmd.Env) != 0 {
		fmt.Fprintf(buf, "\nexport %s\n", shquote.QuoteList(cmd.Env))
	}
	fmt.Fprintf(buf, "\n${exec:-%s} %s \"$@\"\n", exec, shquote.QuoteList(cmdline))
	err = buf.Flush()
	e.Exit(err)
}
