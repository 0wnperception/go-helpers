package yamlSaver

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func Save(params interface{}, path string) (err error) {
	data, err := yaml.Marshal(params)
	if err != nil {
		return
	}
	return ioutil.WriteFile(path, data, 0644)
}

func Load(params interface{}, path string) (err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return yaml.Unmarshal([]byte(data), params)
}
