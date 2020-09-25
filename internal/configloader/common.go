package configloader

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jodydadescott/kerberos-bridge/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	maxIdleConnections int = 2
	requestTimeout     int = 60
)

// GetConfigs ...
func GetConfigs() (*config.Config, error) {

	bootConfigLocation := viper.GetString("config")
	runtimeConfigStringSource := "arg"

	if bootConfigLocation == "" {
		tmp, err := GetRuntimeConfigString()
		if err != nil {
			return nil, err
		}
		bootConfigLocation = tmp
		runtimeConfigStringSource = "system"
	}

	if bootConfigLocation == "" {
		return nil, errors.New("runtime config string not found")
	}

	fmt.Fprintln(os.Stderr, fmt.Sprintf("Using runtime config string %s from %s", bootConfigLocation, runtimeConfigStringSource))

	if strings.HasPrefix(bootConfigLocation, "https://") || strings.HasPrefix(bootConfigLocation, "http://") {
		return getConfigsFromURI(bootConfigLocation)
	}

	return getConfigsFromFile(bootConfigLocation)
}

func getConfigsFromFile(filename string) (*config.Config, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return configsFromBytes(content)
}

func getConfigsFromURI(uri string) (*config.Config, error) {

	fmt.Fprintln(os.Stderr, fmt.Sprintf("Getting config from %s", uri))

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := getHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(fmt.Sprintf("%s returned status code %d", uri, resp.StatusCode))
	}

	config, err := configsFromBytes(b)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func configsFromBytes(input []byte) (*config.Config, error) {

	var config *config.Config

	err := yaml.Unmarshal(input, &config)
	if err == nil {
		return config, nil
	}

	err = json.Unmarshal(input, &config)
	if err == nil {
		return config, nil
	}

	return nil, fmt.Errorf("Config is not valid json or yaml")
}

func getHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
		},
		Timeout: time.Duration(requestTimeout) * time.Second,
	}
}
