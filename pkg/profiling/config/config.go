package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Configuration struct {
	RESTConfig *rest.Config
	Client     client.Client
	Scheme     *runtime.Scheme
}

func (c *Configuration) Load() error {
	// creates the in-cluster config
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	sch := scheme.Scheme
	cl, err := client.New(restConfig, client.Options{
		Scheme: sch,
	})
	if err != nil {
		return err
	}

	c.Scheme = scheme.Scheme
	c.Client = cl
	c.RESTConfig = restConfig

	return nil
}

type config struct {
	Disabled bool `yaml:"disabled"`
}

func GetConfig(path string) (*config, error) {
	data, err := ioutil.ReadFile(filepath.Join(path, "pprof-config.yaml"))
	if err != nil {
		return nil, err
	}

	cfg := &config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
