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
	upstartStoragePath string = "/etc/init"
	upstartTemplate    string = `# {{ .Conf.Name }} - {{ .Conf.Description }}
#
# {{ .Conf.Description }}.

description     "{{ .Conf.Description }}"

start on runlevel [2345]
stop on runlevel [!2345]

respawn
respawn limit 10 5
umask 022

exec {{ .Conf.Exec }}`
)

func init() {
	if IsUpstart() {
		serviceType = Upstart{}
	}
}

type Upstart struct {
	Conf Config
}

func (self Upstart) New(c Config) Service {
	self.Conf = c
	return self
}

func (self Upstart) Register() (output string, err error, code int) {
	if err = self.createUpstartFile(); err != nil {
		return
	}

	return self.Enable()
}

func (self Upstart) Enable() (output string, err error, code int) {
	if _, err := os.Stat(path.Join(upstartStoragePath, self.Conf.Name+".disabled")); !os.IsNotExist(err) {
		err = os.Rename(self.Conf.Name+".disabled", self.Conf.Name+".conf")
	}
	return execute(self.Conf.Log, "initctl", "reload-configuration")
}

func (self Upstart) Start() (output string, err error, code int) {
	return execute(self.Conf.Log, "initctl", "start", self.Conf.Name)
}

func (self Upstart) Stop() (output string, err error, code int) {
	return execute(self.Conf.Log, "initctl", "stop", self.Conf.Name)
}

func (self Upstart) Status() (output string, err error, code int) {
	return execute(self.Conf.Log, "initctl", "status", self.Conf.Name)
}

func (self Upstart) Disable() (output string, err error, code int) {
	err = os.Rename(self.Conf.Name+".conf", self.Conf.Name+".disabled")
	return
}

func (self Upstart) Delete() (output string, err error, code int) {
	os.Remove(path.Join(upstartStoragePath, self.Conf.Name+".conf"))
	os.Remove(path.Join(upstartStoragePath, self.Conf.Name+".disabled"))

	return execute(self.Conf.Log, "initctl", "reload-configuration")
}

func (self Upstart) Run() {
	// To fit Service interface
}

func (self Upstart) IsAnInteractiveSession() (bool, error) {
	// To fit Service interface
	return false, nil
}

func (self Upstart) createUpstartFile() (err error) {
	var b bytes.Buffer
	unitString := bufio.NewWriter(&b)

	unitTmpl, err := template.New("unit").Parse(upstartTemplate)
	if err != nil {
		return fmt.Errorf("createUpstartFile: %s", err)
	}

	err = unitTmpl.Execute(unitString, self)
	if err != nil {
		return fmt.Errorf("createUpstartFile: %s", err)
	}
	unitString.Flush()

	unitPath := path.Join(upstartStoragePath, self.Conf.Name+".conf")

	err = ioutil.WriteFile(unitPath, b.Bytes(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("createUpstartFile: %s", err)
	}

	return nil
}
