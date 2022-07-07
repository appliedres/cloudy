package testutil

import (
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
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Comment
			continue
		}
		if len(trimmed) == 0 {
			continue
		}

		index := strings.Index(trimmed, "=")
		if index > 0 {
			k := trimmed[0:index]
			v := trimmed[index+1:]

			name := cloudy.NormalizeEnvName(k)
			os.Setenv(name, v)
		}

	}
	return nil
}
