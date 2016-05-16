package main

import (
    "os/exec"
    "strings"
    "path/filepath"
    "fmt"
    "runtime"
)

const (
    initbin string = "/sbin/init"
)

func isSysV() bool {
    hcmd := exec.Command(initbin, "--help")
    _, herr := hcmd.CombinedOutput()

    vcmd := exec.Command(initbin, "--version")
    _, verr := vcmd.CombinedOutput()

    if herr != nil && verr != nil {
        return true
    }

    return false
}

func isUpstart() bool {
    vcmd := exec.Command(initbin, "--version")
    output, verr := vcmd.CombinedOutput()

    if verr != nil {
        return false
    }

    return strings.Contains(string(output), "upstart")
}

func isSystemd() bool {
    evaled, err := filepath.EvalSymlinks(initbin)
    if err != nil {
        return false
    }

    return strings.Contains(string(evaled), "systemd")
}

func register() bool {
    if runtime.GOOS == "windows" {
        fmt.Println("windows")
        s := WindowsService{
            Name: "MyTest4",
            BinPath: "\"M:\\Joomla\\GoPath\\src\\SlackService\\slackservice.exe\" --config=\"M:\\Joomla\\GoPath\\src\\SlackService\\config.json\"",
        }
        //fmt.Println(s.Delete())
        fmt.Println(s.Register())
    }
    if (isSysV()) {
        fmt.Println("sysv")
    }
    if (isUpstart()) {
        fmt.Println("upstart")
    }
    if (isSystemd()) {
        fmt.Println("systemd")
        s := SystemD{
            Name: "MyTest",
            Description: "My some test",
            TimeoutStartSec: "1",
            ExecStart: "/root/slackservice.386 --config /root/config.json",
            WantedBy: "multi-user.target",
        }

        fmt.Println(s.Register())
    }

    return true
}

func main() {
    register()
}
