// +build linux

package initme

import (
	"os/exec"
	"syscall"
	"bytes"
	"bufio"
	"fmt"
	"path"
	"io/ioutil"
	"os"
	"text/template"
)

const (
	upstartStoragePath string = "/etc/init"
	upstartTemplate string =`# {{ .Conf.Name }} - {{ .Conf.Description }}
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

func (self Upstart) Register() (output string, err error, code int)  {
	if err = self.createUpstartFile(); err != nil {
        return
    }

    return self.Enable()
}

func (self Upstart) Enable() (output string, err error, code int)  {
	return self.execute("initctl", "reload-configuration")
}

func (self Upstart) Start() (output string, err error, code int)  {
	return self.execute("service", self.Conf.Name, "start")
}

func (self Upstart) Stop() (output string, err error, code int)  {
    return self.execute("service", self.Conf.Name, "stop")
}

func (self Upstart) Status() (output string, err error, code int)  {
    return self.execute("service", self.Conf.Name, "status")
}

func (self Upstart) Disable() (output string, err error, code int)  {
    return self.execute("update-rc.d", self.Conf.Name, "disable", "2", "3", "4", "5")
}

func (self Upstart) Delete() (output string, err error, code int) {
    if _, err := os.Stat(path.Join(sysVstoragePath, self.Conf.Name)); !os.IsNotExist(err) {
        err = os.Remove(path.Join(sysVstoragePath, self.Conf.Name))
    }

	return self.execute("initctl", "reload-configuration")
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

	unitPath := path.Join(upstartStoragePath, self.Conf.Name)

	err = ioutil.WriteFile(unitPath, b.Bytes(), os.ModePerm)
	if err != nil {
        return fmt.Errorf("createUpstartFile: %s", err)
	}

    return nil
}

func (self Upstart) execute(command string, args... string) (output string, err error, code int) {
	self.Conf.Log.Printf("%s %s", command, args)

    cmd := exec.Command(command, args...)
    var waitStatus syscall.WaitStatus
    var outputBytes []byte
    if outputBytes, err = cmd.CombinedOutput(); err != nil {
        // Did the command fail because of an unsuccessful exit code
        if exitError, ok := err.(*exec.ExitError); ok {
            waitStatus = exitError.Sys().(syscall.WaitStatus)
            code = waitStatus.ExitStatus()
        }
    } else {
        // Command was successful
        waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
        code = waitStatus.ExitStatus()
    }

    output = string(outputBytes)
	self.Conf.Log.Println(output, err, code)
    return
}

