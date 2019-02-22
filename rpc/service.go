package rpc

import (
	"fmt"
	"os/exec"
)

type GitRPC interface {
	RunCmd(GitCmd) ([]byte, error)
}

type GitCmd struct {
	// Cmd  string
	Dir  string
	Args []string
}

type Service struct {
	Bin string
}

func NewService() *Service {
	s := &Service{
		Bin: "/usr/bin/git",
	}
	return s
}

// this is for a struct pointer so
// to impl GitRPC you must use &Service
func (s *Service) RunCmd(gc GitCmd) ([]byte, error) {
	fmt.Println("i'm a instance impl GitRPC")
	cmd := exec.Command(s.Bin, gc.Args...)
	cmd.Dir = gc.Dir
	return cmd.Output()
}
