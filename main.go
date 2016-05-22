package initme

import (
    "os/exec"
    "strings"
    "path/filepath"
    "log"
)

const (
    initbin string = "/sbin/init"
)

var serviceType Service

func New(c Config) Service {
    return serviceType.New(c)
}

func IsSysV() bool {
    hcmd := exec.Command(initbin, "--help")
    _, herr := hcmd.CombinedOutput()

    vcmd := exec.Command(initbin, "--version")
    _, verr := vcmd.CombinedOutput()

    if herr != nil && verr != nil {
        return true
    }

    return false
}

func IsUpstart() bool {
    vcmd := exec.Command(initbin, "--version")
    output, verr := vcmd.CombinedOutput()

    if verr != nil {
        return false
    }

    return strings.Contains(string(output), "upstart")
}

func IsSystemD() bool {
    evaled, err := filepath.EvalSymlinks(initbin)
    if err != nil {
        return false
    }

    return strings.Contains(string(evaled), "systemd")
}

type Config struct {
    Name string
    Log *log.Logger
    Description string
    // SysV specific
    Command string
    Provides string
    Required string

    // SystemD specific
    TimeoutStartSec string
    ExecStart string
    WantedBy string

    // Windows Specific
    Job func()
    Type string
    StartType string
    Error string
    BinPath string
    Group string
    Tag string
    Depend string
    Obj string
    DisplayName string
    Password string
}

type Service interface {

    New(Config) Service

    Register() (output string, err error, code int)

    Start() (output string, err error, code int)

    Stop() (output string, err error, code int)

    Status() (output string, err error, code int)

    Disable() (output string, err error, code int)

    Delete() (output string, err error, code int)

    Run()

    IsAnInteractiveSession() (bool, error)
}
