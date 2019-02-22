package main

/**
This a git server with go implemention
which implements https://mirrors.edge.kernel.org/pub/software/scm/git/docs/technical/http-protocol.html
*/

/**
Useful links:
http://scottchacon.com/2010/03/04/smart-http.html
https://mirrors.edge.kernel.org/pub/software/scm/git/docs/git-http-backend.html
*/

/**
For now i only try to impl the smart protocal, so
the git client's version must be over 1.6.6
*/

/**
tested server side git version is 2.16.5
*/

import (
	"flag"
	"fmt"

	"deoops/git-http-backend/conf"
	"deoops/git-http-backend/web"
)

func init() {
	flag.StringVar(&conf.Default.Home, "b", conf.Default.Home, "goty home data dir")
}

func main() {
	fmt.Println("start my own server")
	flag.Parse()

	web.Run()
}
