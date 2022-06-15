package yamlSaver

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func Save(params interface{}, path string) (err error) {
	data, err := yaml.Marshal(params)
	if err != nil {
		return
	}

	if err = ioutil.WriteFile(path, data, 0644); err != nil {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		_, err = f.Write(data)
		f.Close()
		return err
	}
	return
}

func Load(params interface{}, path string) (err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	return yaml.Unmarshal([]byte(data), params)
}
