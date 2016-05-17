package initme

import (
    "bufio"
    "os"
    "path"
    "text/template"
    "io/ioutil"
    "os/exec"
    "fmt"
    "syscall"
    "bytes"
	"log"
)

const (
    unitStoragePath string  = "/etc/systemd/system"
    unitTemplate string = `[Unit]
Description={{ .Description }}
After=network.target

[Service]
TimeoutStartSec={{ .TimeoutStartSec }}
ExecStart={{ .ExecStart }}

[Install]
WantedBy={{ .WantedBy }}`
)

type SystemD struct {
    Name string
	Log *log.Logger
    Description string
    TimeoutStartSec string
    ExecStart string
    WantedBy string
}

func (self SystemD) Register() (output string, err error, code int)  {
    if err = self.createUnitFile(); err != nil {
        return
    }

    return self.Enable()
}

func (self SystemD) Status() (output string, err error, code int) {
    return self.execute("status", self.Name + ".service")
}

func (self SystemD) Enable() (output string, err error, code int) {
    return self.execute("enable", path.Join(unitStoragePath, self.Name + ".service"))
}

func (self SystemD) Disable() (err error) {
    cmd := exec.Command("systemctl", "disable", self.Name + ".service")
    _, err = cmd.CombinedOutput()

    if err != nil {
        return fmt.Errorf("Disable: %s", err)
	}

    return nil
}

func (self SystemD) Delete() (err error) {
	if _, err := os.Stat(path.Join(unitStoragePath, self.Name + ".service")); os.IsNotExist(err) {
		return
	}

	err = os.Remove(path.Join(unitStoragePath, self.Name + ".service"))
    return
}

func (self SystemD) Start() (output string, err error, code int)  {
    return self.execute("start", self.Name + ".service")
}

func (self SystemD) Stop() (err error) {
    cmd := exec.Command("systemctl", "stop", self.Name + ".service")
    _, err = cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("Stop: %s", err)
	}

    return nil
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

	unitPath := path.Join(unitStoragePath, self.Name + ".service")

	err = ioutil.WriteFile(unitPath, b.Bytes(), os.ModePerm)
	if err != nil {
        return fmt.Errorf("createUnitFile: %s", err)
	}

    return nil
}

func (self SystemD) execute(args... string) (output string, err error, code int) {
    cmd := exec.Command("systemctl", args...)
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

    return string(outputBytes), err, code
}