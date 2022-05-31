package cloudy

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func StartDocker(name string, args []string, waitFor string) (bool, error) {
	fmt.Println("Starting instance in docker for testing")
	cmdArgs := []string{"run", "--rm", "--name", name}
	cmdArgs = append(cmdArgs, args...)

	// cmd := exec.Command("docker", "run", "--rm", "--name", "cloudy-test-elasticsearch", "-e", "discovery.type=single-node", "-d", "-p", "9201:9200", "elasticsearch:7.14.2")
	cmd := exec.Command("docker", cmdArgs...)

	var out bytes.Buffer
	var errs bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &errs
	err := cmd.Run()
	started := true
	if err != nil {
		errstr := errs.String()
		if strings.HasPrefix(errstr, "docker: Error response from daemon: Conflict.") {
			fmt.Println("Already running")
			started = false
		} else {
			fmt.Println(out.String())
			fmt.Println(errs.String())
			return started, err
		}
	}

	// check to see if available
	if waitFor != "" {
		fmt.Printf("Waiting for %v to become avialable\n", waitFor)
		found := WaitForAddress(waitFor, 60*time.Second)
		if !found {
			return started, errors.New("unable to connect")
		}
	} else {
		fmt.Printf("Waiting 30 sec\n")
		time.Sleep(30 * time.Second)
	}

	fmt.Println("Completed Startup")
	return started, nil
}

func ShutdownDocker(name string) error {
	fmt.Printf("Shutting down %v\n", name)

	cmd := exec.Command("docker", "stop", name)
	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Printf("Completed %v Shutdown\n", name)

	return nil
}
