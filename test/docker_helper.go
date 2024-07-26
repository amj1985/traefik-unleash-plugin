package testmain

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"testing"
)

const (
	errorExitCode = 1
)

func Setup(m *testing.M) {
	exitCode := errorExitCode
	tc, _ := compose.NewDockerCompose("../../docker-compose.yml")

	defer func() {
		fmt.Printf("%v\n", exitCode)
		os.Exit(exitCode)
	}()

	defer func() {
		fmt.Println("docker-compose down")
		_ = tc.Down(context.Background())
	}()

	err := tc.
		WithEnv(map[string]string{"TESTCONTAINERS_RYUK_DISABLED ": "true"}).
		WaitForService("traefik", wait.ForHTTP("/foo")).
		Up(context.Background())

	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}
	exitCode = m.Run()
}
