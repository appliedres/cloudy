package cloudy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

var _ EnvironmentService = (*FileEnvironmentService)(nil)

type FileEnvironmentService struct {
	Filename string
	loaded   bool
	lock     sync.RWMutex
	data     map[string]string
}

func NewFileEnvironmentService(filename string) *FileEnvironmentService {
	return &FileEnvironmentService{
		Filename: filename,
		data:     make(map[string]string),
	}
}

func (fs *FileEnvironmentService) GetAll() map[string]string {
	return fs.data
}

func (fs *FileEnvironmentService) GetMap() map[string]string {
	return fs.data
}

func (fs *FileEnvironmentService) loadIfNeeded() error {
	if fs.loaded {
		return nil
	}
	return fs.Load(false)
}

func (fs *FileEnvironmentService) Get(name string) (string, error) {
	fs.loadIfNeeded()

	fs.lock.Lock()
	defer fs.lock.Unlock()
	key := ToEnvName(name, "")
	rtn := fs.data[key]
	if rtn != "" {
		return rtn, nil
	}

	rtn = fs.data[name]
	return rtn, nil
}

func (fs *FileEnvironmentService) Set(name string, value string) error {
	fs.loadIfNeeded()

	fs.lock.Lock()
	defer fs.lock.Unlock()

	key := ToEnvName(name, "")
	fs.data[key] = value

	return fs.Save()
}

func (fs *FileEnvironmentService) SetMany(many map[string]string) error {
	fs.loadIfNeeded()

	fs.lock.Lock()
	defer fs.lock.Unlock()

	for k, value := range many {
		fs.data[k] = value
	}

	return fs.Save()
}
func (fs *FileEnvironmentService) Load(mustExist bool) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	fs.loaded = true
	exists, err := Exists(fs.Filename)
	if err != nil && mustExist {
		return err
	}
	if !exists {
		if mustExist {
			return fmt.Errorf("%v does not exist", fs.Filename)
		}
		return nil
	}

	if strings.HasSuffix(fs.Filename, ".json") {
		return fs.loadFromJson()
	}
	return fs.loadFromEnv()
}

func (fs *FileEnvironmentService) Save() error {
	if strings.HasSuffix(fs.Filename, ".json") {
		return fs.saveAsJson()
	}
	return fs.saveAsEnv()
}

func LoadEnvFromString(data string) (map[string]string, error) {
	rtn := make(map[string]string)
	lines := strings.Split(data, "\n")
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

			name := NormalizeKey(k)
			rtn[name] = v
		}

	}
	return rtn, nil
}

func (fs *FileEnvironmentService) saveAsJson() error {
	data, err := json.MarshalIndent(fs.data, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.Filename, data, 0600)
}

func (fs *FileEnvironmentService) saveAsEnv() error {
	var sb strings.Builder
	for k, v := range fs.data {
		key := ToEnvName(k, "")
		sb.WriteString(fmt.Sprintf("%v=%v\n", key, v))
	}
	return os.WriteFile(fs.Filename, []byte(sb.String()), 0600)
}

func (fs *FileEnvironmentService) loadFromJson() error {
	data, err := os.ReadFile(fs.Filename)
	if err != nil {
		return err
	}

	var src map[string]interface{}
	err = json.Unmarshal(data, &src)
	if err != nil {
		return err
	}

	mapStr := make(map[string]string)
	for k, v := range src {
		val := fmt.Sprintf("%v", v)
		mapStr[k] = val
	}
	fs.data = mapStr
	return nil
}

func (fs *FileEnvironmentService) loadFromEnv() error {
	data, err := os.ReadFile(fs.Filename)
	if err != nil {
		return err
	}

	m, err := LoadEnvFromString(string(data))
	if err != nil {
		return err
	}
	fs.data = m
	return nil
}

func NormalizeKey(key string) string {
	// Lowercase
	// hypens to dots
	// underscores to dots
	rtn := strings.ToLower(key)
	rtn = strings.ReplaceAll(rtn, "-", ".")
	rtn = strings.ReplaceAll(rtn, "_", ".")

	return rtn
}
