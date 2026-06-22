package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	_getGRPCGen = "go get -u github.com/gogo/protobuf/protoc-gen-gofast"
	_grpcProtoc = "protoc --proto_path=%s --proto_path=%s --proto_path=%s --gofast_out=plugins=grpc," +
		"Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types," +
		"Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:."
)

func installGRPCGen() error {
	if _, err := exec.LookPath("protoc-gen-gofast"); err != nil {
		if err := goget(_getGRPCGen); err != nil {
			return err
		}
	}
	return nil
}

// findProjectRoot 从 start 目录开始向上查找，直到找到包含 go.mod 的目录
func findProjectRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found starting from %s", start)
		}
		dir = parent
	}
}

func genGRPC(files []string) error {
	pwd, _ := os.Getwd()
	gosrc := path.Join(gopath(), "src")
	ext, err := latestKratos()
	if err != nil {
		return err
	}

	// 以 go.mod 所在目录作为项目根
	projectRoot, err := findProjectRoot(pwd)
	if err != nil {
		return err
	}

	// 项目名和项目根的父目录
	projectName := filepath.Base(projectRoot)
	parentOfProject := filepath.Dir(projectRoot)

	// proto 文件相对于项目根的路径
	relPwd, err := filepath.Rel(projectRoot, pwd)
	if err != nil {
		return err
	}
	// 让 RegisterFile 带上项目名前缀
	var cmdFiles []string
	for _, file := range files {
		cmdFiles = append(cmdFiles, filepath.Join(projectName, relPwd, file))
	}

	line := fmt.Sprintf(_grpcProtoc, gosrc, ext, parentOfProject)
	log.Println(line, strings.Join(cmdFiles, " "))
	args := strings.Split(line, " ")
	args = append(args, cmdFiles...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = parentOfProject
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
