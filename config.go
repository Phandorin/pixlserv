package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

const (
	LRU = "LRU"
	LFU = "LFU"
)

const (
	defaultThrottlingRate             = 60              // Requests per min
	defaultCacheLimit                 = 0               // No. of bytes
	defaultUploadMaxFileSize          = 5 * 1024 * 1024 // No. of bytes
	defaultAllowCustomTransformations = true
	defaultAllowCustomScale           = true
	defaultAsyncUploads               = false
	defaultAuthorisedGet              = false
	defaultAuthorisedUpload           = false
	defaultLocalPath                  = "local-images"
	defaultCacheStrategy              = LRU
)

var (
	Config Configuration
)

type Configuration struct {
	throttlingRate, cacheLimit, uploadMaxFileSize                                               int
	allowCustomTransformations, allowCustomScale, asyncUploads, authorisedGet, authorisedUpload bool
	localPath, cacheStrategy                                                                    string
	transformations                                                                             map[string]Params
	eagerTransformations                                                                        []Params
}

func configInit(configFilePath string) error {
	Config = Configuration{defaultThrottlingRate, defaultCacheLimit, defaultUploadMaxFileSize, defaultAllowCustomTransformations, defaultAllowCustomScale, defaultAsyncUploads, defaultAuthorisedGet, defaultAuthorisedUpload, defaultLocalPath, defaultCacheStrategy, make(map[string]Params), make([]Params, 0)}

	if configFilePath == "" {
		return nil
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)

	throttlingRate, ok := m["throttling-rate"].(int)
	if ok && throttlingRate >= 0 {
		Config.throttlingRate = throttlingRate
	}

	uploadMaxFileSize, ok := m["upload-max-file-size"].(int)
	if ok && uploadMaxFileSize > 0 {
		Config.uploadMaxFileSize = uploadMaxFileSize
	}

	allowCustomTransformations, ok := m["allow-custom-transformations"].(bool)
	if ok {
		Config.allowCustomTransformations = allowCustomTransformations
	}

	allowCustomScale, ok := m["allow-custom-scale"].(bool)
	if ok {
		Config.allowCustomScale = allowCustomScale
	}

	asyncUploads, ok := m["async-uploads"].(bool)
	if ok {
		Config.asyncUploads = asyncUploads
	}

	authorisation, ok := m["authorisation"].(map[interface{}]interface{})
	if ok {
		get, ok := authorisation["get"].(bool)
		if ok {
			Config.authorisedGet = get
		}
		upload, ok := authorisation["upload"].(bool)
		if ok {
			Config.authorisedUpload = upload
		}
	}

	localPath, ok := m["local-path"].(string)
	if ok {
		Config.localPath = localPath
	}

	cache, ok := m["cache"].(map[interface{}]interface{})
	if ok {
		limit, ok := cache["limit"].(int)
		if ok {
			Config.cacheLimit = limit
		}

		strategy, ok := cache["strategy"].(string)
		if ok && (strategy == LRU || strategy == LFU) {
			Config.cacheStrategy = strategy
		}
	}

	transformations, ok := m["transformations"].([]interface{})
	if ok {
		for _, transformationInterface := range transformations {
			transformation, ok := transformationInterface.(map[interface{}]interface{})
			if ok {
				parametersStr, ok := transformation["parameters"].(string)
				if ok {
					params, err := parseParameters(parametersStr)
					if err != nil {
						return fmt.Errorf("invalid transformation parameters: %s (%s)", parametersStr, err)
					}
					name, ok := transformation["name"].(string)
					if ok {
						Config.transformations[name] = params
						eager, ok := transformation["eager"].(bool)

						if ok && eager {
							Config.eagerTransformations = append(Config.eagerTransformations, params)
						}
					}
				}
			}
		}
	}

	return nil
}
