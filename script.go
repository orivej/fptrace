package main

import (
	"bufio"
	"fmt"
	"os"
	"path"

	"github.com/djmitche/shquote"
	"github.com/orivej/e"
)

func writeScript(dir string, r Record) {
	name := fmt.Sprintf("%d-%d-%s", r.Cmd.Parent, r.Cmd.ID, path.Base(r.Cmd.Path))
	f, err := os.OpenFile(path.Join(dir, name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777) //#nosec
	e.Exit(err)
	defer e.CloseOrPrint(f)

	buf := bufio.NewWriter(f)
	fmt.Fprintln(buf, "#!/bin/sh")
	if len(r.Cmd.Env) != 0 {
		fmt.Fprintf(buf, "\nexport %s\n", shquote.QuoteList(r.Cmd.Env))
	}
	fmt.Fprintf(buf, "\ncd %s\n", shquote.Quote(r.Cmd.Dir))
	fmt.Fprintf(buf, "\n${exec:-exec} %s \"$@\"\n", shquote.QuoteList(r.Cmd.Args))
	err = buf.Flush()
	e.Exit(err)
}
