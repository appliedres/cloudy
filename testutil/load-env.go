package testutil

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/appliedres/cloudy"
)

func CreateTestEnvMgr() *cloudy.EnvManager {
	MustSetTestEnv()
	em := cloudy.GetDefaultEnvManager()
	em.LoadSources("test")
	return em
}

// Starts in the current directory and checks for "arkloud.env" OR "arkloud-conf/arkloud.env"
// Keeps going up until it either finds it or there are no more directories
func MustSetTestEnv() {
	// First look for "ARKLOUD_ENV_CI"
	mp := cloudy.NewCIEnvironmentService()
	if mp != nil {
		return
	}

	// Now Find the first file
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("IO Error : %v", err)
	}
	working := dir
	for {
		// Check
		path := filepath.Join(working, "arkloud.env")
		exists, _ := cloudy.Exists(path)
		if exists {
			os.Setenv("ARKLOUD_ENVFILE", path)
			return
		}

		// Check
		path = filepath.Join(working, "arkloud-conf", "arkloud.env")
		exists, _ = cloudy.Exists(path)
		if exists {
			os.Setenv("ARKLOUD_ENVFILE", path)
			return
		}

		// next
		working = filepath.Dir(working)
		if working == "" {
			log.Fatal("NOT FOUND")
		}
	}
}

func EnvFileMustExist(path string) {
	exists, _ := cloudy.Exists("../../arkloud-conf/arkloud.env")
	if !exists {
		log.Fatalf("CREATE FILE: %v", path)
	}
	os.Setenv("ARKLOUD_ENVFILE", path)
}

// Loads a file delimited by \n and =
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
