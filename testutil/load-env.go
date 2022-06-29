package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/appliedres/cloudy"
)

//Loads a file delimited by \n and =
func LoadEnv(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	all := string(data)
	lines := strings.Split(all, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Comment
			continue
		}
		if len(trimmed) == 0 {
			continue
		}

		parts := strings.Split(trimmed, "=")
		if len(parts) == 1 {
			return fmt.Errorf("invalid line %v: %v", i, line)
		}

		name := cloudy.NormalizeEnvName(parts[0])
		value := parts[1]

		os.Setenv(name, value)
	}
	return nil
}
