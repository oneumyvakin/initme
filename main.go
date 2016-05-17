package initme

import (
    "os/exec"
    "strings"
    "path/filepath"
)

const (
    initbin string = "/sbin/init"
)

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

