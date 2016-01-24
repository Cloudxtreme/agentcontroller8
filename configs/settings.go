package configs

import (
	"io/ioutil"
	"os"

	"fmt"
	"github.com/naoina/toml"
	"net"
)

//HTTPBinding defines the address that should be bound on and optional tls certificates
type HTTPBinding struct {
	Address string
	TLS     []struct {
		Cert string
		Key  string
	}
	ClientCA []struct {
		Cert string
	}
}

type Extension struct {
	Enabled      bool
	Module       string
	PythonBinary string
	PythonPath   string
	Settings     map[string]string
}

func (e *Extension) GetPythonBinary() string {
	if e.PythonBinary != "" {
		return e.PythonBinary
	}

	return "python"
}

type validationError struct {
	base    error
	message string
}

func (v *validationError) Error() string {
	return fmt.Sprintf("%s: %s", v.message, v.base)
}

func ValidationError(message string, base error) error {
	return &validationError{base, message}
}

//Settings are the configurable options for the AgentController
type Settings struct {
	Main struct {
		RedisHost     string
		RedisPassword string
	}

	Listen []HTTPBinding

	Influxdb struct {
		Host     string
		Db       string
		User     string
		Password string
	}

	Events      Extension
	Processor   Extension
	Jumpscripts Extension

	Syncthing struct {
		Port int
	}
}

func (s *Settings) Validate() []error {
	errors := make([]error, 0)
	//1- Validate TCP Addr of redis.
	if redisAddress, err := net.ResolveTCPAddr("tcp", s.Main.RedisHost); err != nil {
		errors = append(errors, ValidationError("[main] `redis_host` error", err))
	} else if redisAddress.IP.String() == "" || redisAddress.Port == 0 {
		errors = append(errors, ValidationError("[main'] `redis_host` error", fmt.Errorf("Invalid address :%s", redisAddress)))
	}

	for i, tcpBind := range s.Listen {
		if address, err := net.ResolveTCPAddr("tcp", tcpBind.Address); err != nil {
			errors = append(errors, ValidationError(fmt.Sprintf("[listen] [%d] `address` error", i), err))
		} else if address.Port == 0 {
			errors = append(errors, ValidationError(fmt.Sprintf("[listen] [%d] `address` error", i), fmt.Errorf("Invalid address :%s", address)))
		}
	}

	return errors
}

//LoadSettingsFromTomlFile does exactly what the name says, it loads a toml in a Settings struct
func LoadSettingsFromTomlFile(filename string) (*Settings, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var settings Settings
	err = toml.Unmarshal(buf, &settings)

	return &settings, nil
}

//TLSEnabled returns true if TLS settings are present
func (httpBinding HTTPBinding) TLSEnabled() bool {
	return len(httpBinding.TLS) > 0
}

//ClientCertificateRequired returns true if ClientCA's are present
func (httpBinding HTTPBinding) ClientCertificateRequired() bool {
	return len(httpBinding.ClientCA) > 0
}
