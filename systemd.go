// +build linux

package initme

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"
)

const (
	unitStoragePath string = "/etc/systemd/system"
	unitTemplate    string = `[Unit]
Description={{ .Conf.Description }}
After=network.target

[Service]
TimeoutStartSec={{ .Conf.TimeoutStartSec }}
ExecStart={{ .Conf.ExecStart }}

[Install]
WantedBy={{ .Conf.WantedBy }}`
)

func init() {
	if IsSystemD() {
		serviceType = SystemD{}
	}
}

type SystemD struct {
	Conf Config
}

func (self SystemD) New(c Config) Service {

	self.Conf = c

	return self
}

func (self SystemD) Register() (output string, err error, code int) {
	if err = self.createUnitFile(); err != nil {
		return
	}

	return self.Enable()
}

func (self SystemD) Start() (output string, err error, code int) {
	return execute(self.Conf.Log, "systemctl", "start", self.Conf.Name+".service")
}

func (self SystemD) Stop() (output string, err error, code int) {
	return execute(self.Conf.Log, "systemctl", "stop", self.Conf.Name+".service")
}

func (self SystemD) Status() (output string, err error, code int) {
	return execute(self.Conf.Log, "systemctl", "status", self.Conf.Name+".service")
}

func (self SystemD) Enable() (output string, err error, code int) {
	return execute(self.Conf.Log, "systemctl", "enable", path.Join(unitStoragePath, self.Conf.Name+".service"))
}

func (self SystemD) Disable() (output string, err error, code int) {
	return execute(self.Conf.Log, "systemctl", "disable", self.Conf.Name+".service")
}

func (self SystemD) Delete() (output string, err error, code int) {
	if _, err := os.Stat(path.Join(unitStoragePath, self.Conf.Name+".service")); os.IsNotExist(err) {
		return output, nil, code
	}

	err = os.Remove(path.Join(unitStoragePath, self.Conf.Name+".service"))
	return
}

func (self SystemD) Run() {
	// To fit Service interface
}

func (self SystemD) createUnitFile() (err error) {
	var b bytes.Buffer
	unitString := bufio.NewWriter(&b)

	unitTmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return fmt.Errorf("createUnitFile: %s", err)
	}

	err = unitTmpl.Execute(unitString, self)
	if err != nil {
		return fmt.Errorf("createUnitFile: %s", err)
	}
	unitString.Flush()

	unitPath := path.Join(unitStoragePath, self.Conf.Name+".service")

	err = ioutil.WriteFile(unitPath, b.Bytes(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("createUnitFile: %s", err)
	}

	return nil
}

func (self SystemD) IsAnInteractiveSession() (bool, error) {
	// To fit Service interface
	return false, nil
}
