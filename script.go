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

	buf := bufio.NewWriter(f)
	fmt.Fprintln(buf, "#!/bin/sh")
	fmt.Fprintf(buf, "\ncd %s\n", shquote.Quote(cmd.Dir))
	if len(cmd.Env) != 0 {
		fmt.Fprintf(buf, "\nexport %s\n", shquote.QuoteList(cmd.Env))
	}
	fmt.Fprintf(buf, "\n${exec:-exec} %s \"$@\"\n", shquote.QuoteList(cmd.Args))
	err = buf.Flush()
	e.Exit(err)
}
