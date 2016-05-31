// +build linux

package initme

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	sysVstoragePath string = "/etc/init.d"
	sysVtemplate    string = `#!/bin/sh
### BEGIN INIT INFO
# Provides:          {{ .Conf.Provides }}
# Required-Start:    {{ .Conf.Required }}
# Required-Stop:     $null
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: {{ .Conf.Description }}
# Description:       {{ .Conf.Description }}.
### END INIT INFO


cmd="{{ .Conf.Command }}"

name=` + "`" + "basename $0" + "`" + `
pid_file="/var/run/$name.pid"
stdout_log="/var/log/$name.log"
stderr_log="/var/log/$name.err"

get_pid() {
    cat "$pid_file"
}

is_running() {
    [ -f "$pid_file" ] && ps ` + "`" + "get_pid" + "`" + ` > /dev/null 2>&1
}

case "$1" in
    start)
    if is_running; then
        echo "Already started"
    else
        echo "Starting $name"

        $cmd >> "$stdout_log" 2>> "$stderr_log" &

        echo $! > "$pid_file"

        if ! is_running; then
            echo "Unable to start, see $stdout_log and $stderr_log"
            exit 1
        fi
    fi
    ;;
    stop)
    if is_running; then
        echo -n "Stopping $name.."
        kill ` + "`" + "get_pid" + "`" + `
        for i in {1..10}
        do
            if ! is_running; then
                break
            fi

            echo -n "."
            sleep 1
        done
        echo

        if is_running; then
            echo "Not stopped; may still be shutting down or shutdown may have failed"
            exit 1
        else
            echo "Stopped"
            if [ -f "$pid_file" ]; then
                rm "$pid_file"
            fi
        fi
    else
        echo "Not running"
    fi
    ;;
    restart)
    $0 stop
    if is_running; then
        echo "Unable to stop, will not attempt to start"
        exit 1
    fi
    $0 start
    ;;
    status)
    if is_running; then
        echo "Running"
    else
        echo "Stopped"
        exit 1
    fi
    ;;
    *)
    echo "Usage: $0 {start|stop|restart|status}"
    exit 1
    ;;
esac

exit 0`
)

func init() {
	if IsSysV() {
		serviceType = SysV{}
	}
}

type SysV struct {
	Conf Config
}

func (self SysV) New(c Config) Service {

	self.Conf = c

	return self
}

func (self SysV) Register() (output string, err error, code int) {
	if err = self.createServiceFile(); err != nil {
		return
	}

	return self.Enable()
}

func (self SysV) Enable() (output string, err error, code int) {
	return self.initier("enable")
}

func (self SysV) Start() (output string, err error, code int) {
	return execute(self.Conf.Log, path.Join(sysVstoragePath, self.Conf.Name), "start")
}

func (self SysV) Stop() (output string, err error, code int) {
	return execute(self.Conf.Log, path.Join(sysVstoragePath, self.Conf.Name), "stop")
}

func (self SysV) Status() (output string, err error, code int) {
	return execute(self.Conf.Log, path.Join(sysVstoragePath, self.Conf.Name), "status")
}

func (self SysV) Disable() (output string, err error, code int) {
	return self.initier("disable")
}

func (self SysV) Delete() (output string, err error, code int) {
	if _, err := os.Stat(path.Join(sysVstoragePath, self.Conf.Name)); !os.IsNotExist(err) {
		err = os.Remove(path.Join(sysVstoragePath, self.Conf.Name))
	}

	return self.initier("delete")
}

func (self SysV) Run() {
	// To fit Service interface
}

func (self SysV) initier(command string) (output string, err error, code int) {
	var ctl string
	output, err, code = execute(self.Conf.Log, "which", "update-rc.d")
	if err == nil {
		ctl = "update-rc.d"
	}
	output, err, code = execute(self.Conf.Log, "which", "chkconfig")
	if err == nil {
		ctl = "chkconfig"
	}
	if ctl == "" {
		self.Conf.Log.Printf("Control utility not found. Aborting.")
		return
	}

	self.Conf.Log.Printf("Control utility found: " + ctl)

	cmdSet := map[string]map[string]string{
		"update-rc.d": map[string]string{
			"enable":  self.Conf.Name + " defaults",
			"disable": self.Conf.Name + " disable 2 3 4 5",
			"delete":  self.Conf.Name + " remove",
		},
		"chkconfig": map[string]string{
			"enable":  "-s " + self.Conf.Name + " on",
			"disable": "-s " + self.Conf.Name + " off",
			"delete":  "-s " + self.Conf.Name + " off",
		},
	}

	args := strings.Split(cmdSet[ctl][command], " ")

	return execute(self.Conf.Log, ctl, args...)
}

func (self SysV) createServiceFile() (err error) {
	var b bytes.Buffer
	unitString := bufio.NewWriter(&b)

	unitTmpl, err := template.New("unit").Parse(sysVtemplate)
	if err != nil {
		return fmt.Errorf("createServiceFile: %s", err)
	}

	err = unitTmpl.Execute(unitString, self)
	if err != nil {
		return fmt.Errorf("createServiceFile: %s", err)
	}
	unitString.Flush()

	unitPath := path.Join(sysVstoragePath, self.Conf.Name)

	err = ioutil.WriteFile(unitPath, b.Bytes(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("createServiceFile: %s", err)
	}

	return nil
}

func (self SysV) IsAnInteractiveSession() (bool, error) {
	// To fit Service interface
	return false, nil
}
