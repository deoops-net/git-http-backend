package web

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"deoops/git-http-backend/rpc"

	"github.com/gorilla/mux"
)

const BASE = "/tmp/"

// TODO refactor
// here we run a http server with gorilla's mux lib
// this 3 routes are all needed for basic using
// but like grack it impls many more
func Run() {
	r := mux.NewRouter()

	r.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`hello world\n`))
		},
	)

	r.HandleFunc(
		"/{repo}/info/refs",
		HandleInfoRef,
	)

	// git push ...
	// client uploads  and will call server side reciver
	r.HandleFunc(
		"/{repo}/git-receive-pack",
		ServiceRpc,
	)

	// git pull
	// client fetch  and will call server side upload
	r.HandleFunc(
		"/{repo}/git-upload-pack",
		UploadRpc,
	)

	log.Fatal(http.ListenAndServe(":2208", r))
}

func HandleInfoRef(w http.ResponseWriter, r *http.Request) {
	rpcCmd := r.FormValue("service")
	args := []string{strings.Replace(rpcCmd, "git-", "", -1), "--stateless", "--advertise-refs", "."}
	// set no cache header
	w.Header().Set("Expires", "Fri, 01 Jan 1980 00:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	// set rpc header
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", rpcCmd))

	cmd := rpc.GitCmd{
		Dir:  BASE + mux.Vars(r)["repo"],
		Args: args,
	}
	s := rpc.NewService()

	out, err := s.RunCmd(cmd)
	if err != nil {
		log.Println("got error")
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(packetWrite("# service=" + rpcCmd + "\n"))
	w.Write(packetFlush())
	w.Write(out)
}

func packetFlush() []byte {
	return []byte("0000")
}

func packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)

	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s

	}

	return []byte(s + str)
}

// POST /mytest/git-receive-pack HTTP/1.1
func ServiceRpc(w http.ResponseWriter, r *http.Request) {
	if len(r.Header["Authorization"]) == 0 {
		w.Header().Set("WWW-Authenticate", "Basic reaml=Restrited")
		w.WriteHeader(401)
		return
	}
	// check user->repo permission
	payload, err := base64.StdEncoding.DecodeString(
		r.Header["Authorization"][0][len("Basic "):],
	)

	if err != nil {
		w.Header().Set("WWW-Authenticate", "Basic reaml=Restrited")
		w.WriteHeader(401)
		return
	}
	auth := bytes.SplitN(payload, []byte(`:`), 2)
	if string(auth[0]) != "123" && string(auth[1]) != "123" {
		w.Header().Set("WWW-Authenticate", "Basic reaml=Restrited")
		w.WriteHeader(401)
		return
	}

	fmt.Println("auth success")

	// call git commands
	rpcCmd := "git-receive-pack"
	dir := BASE + mux.Vars(r)["repo"]

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-result", rpcCmd))
	w.WriteHeader(http.StatusOK)

	args := []string{"receive-pack", "--stateless-rpc", dir}

	log.Println(args)
	cmd := exec.Command("/usr/bin/git", args...)
	version := r.Header.Get("Git-Protocol")
	if len(version) != 0 {
		cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_PROTOCOL=%s", version))
	}
	cmd.Dir = dir
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Print(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Print(err)
	}

	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)
		defer reader.Close()
	default:
		log.Println("here")
		reader = r.Body
		defer reader.Close()
	}
	io.Copy(in, reader)
	in.Close()
	io.Copy(w, stdout)
	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

func UploadRpc(w http.ResponseWriter, r *http.Request) {
	rpcCmd := "git-upload-pack"
	dir := BASE + mux.Vars(r)["repo"]
	// access := hasAccess(r, dir, rpc, true)

	// if access == false {
	// renderNoAccess(w)
	// return
	// }
	// log.Println("!!")
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-result", rpcCmd))
	w.WriteHeader(http.StatusOK)

	args := []string{"upload-pack", "--stateless-rpc", dir}

	log.Println(args)
	cmd := exec.Command("/usr/bin/git", args...)
	version := r.Header.Get("Git-Protocol")
	if len(version) != 0 {
		cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_PROTOCOL=%s", version))
	}
	cmd.Dir = dir
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Print(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Print(err)
	}

	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)
		defer reader.Close()
	default:
		log.Println("here")
		reader = r.Body
		defer reader.Close()
	}
	io.Copy(in, reader)
	in.Close()
	io.Copy(w, stdout)
	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
