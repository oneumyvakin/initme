package initme

import (
	"log"
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
	sysVstoragePath string  = "/etc/init.d"
	sysVtemplate string = `#!/bin/sh
### BEGIN INIT INFO
# Provides:
# Required-Start:    {{ .Required }}
# Required-Stop:     $null
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: {{ .Description }}
# Description:       {{ .Description }}.
### END INIT INFO


cmd="{{ .Command }}"

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

type SysV struct {
    Name string
	Log *log.Logger
    Description string
    Command string
    Required string
}

func (self SysV) Register() (output string, err error, code int)  {
	if err = self.createServiceFile(); err != nil {
        return
    }

    return self.Enable()
}

func (self SysV) Enable() (output string, err error, code int)  {
	return self.execute("chkconfig", self.Name, "on")
}

func (self SysV) Start() (output string, err error, code int)  {
	return self.execute(path.Join(sysVstoragePath, self.Name), "start")
}

func (self SysV) Stop() (output string, err error, code int)  {
    return self.execute(path.Join(sysVstoragePath, self.Name), "stop")
}

func (self SysV) Status() (output string, err error, code int)  {
    return self.execute(path.Join(sysVstoragePath, self.Name), "status")
}

func (self SysV) Disable() (output string, err error, code int)  {
    return self.execute("chkconfig", self.Name, "off")
}

func (self SysV) Delete() (err error) {
	if _, err := os.Stat(path.Join(sysVstoragePath, self.Name)); os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(path.Join(sysVstoragePath, self.Name))
    return
}

func (self SysV) execute(command string, args... string) (output string, err error, code int) {
	self.Log.Printf("%s %s", command, args)

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
	self.Log.Println(output, err, code)
    return
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

	unitPath := path.Join(sysVstoragePath, self.Name)

	err = ioutil.WriteFile(unitPath, b.Bytes(), os.ModePerm)
	if err != nil {
        return fmt.Errorf("createServiceFile: %s", err)
	}

    return nil
}