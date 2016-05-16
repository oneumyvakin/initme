package initme

import (
    "os/exec"
    "strings"
    "path/filepath"
    "fmt"
    "runtime"

    "golang.org/x/sys/windows/svc"
	"log"
	"os"
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

func IsSystemd() bool {
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

            BinPath: "\"C:\\Users\\oneumyvakin.SWSOFT\\Desktop\\GoPath\\src\\slackservice\\slackservice.exe\" --config \"C:\\Users\\oneumyvakin.SWSOFT\\Desktop\\GoPath\\src\\slackservice\\config.json\"",
        }

		for _, arg := range os.Args {
			if arg == "--install" {
        		fmt.Println(s.Register())
			}
		}

		isIntSess, err := svc.IsAnInteractiveSession()
		if err != nil {
			log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
		}
		if !isIntSess {
			s.Run()
			return true
		}



    }
    if (IsSysV()) {
        fmt.Println("sysv")
    }
    if (IsUpstart()) {
        fmt.Println("upstart")
    }
    if (IsSystemd()) {
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
