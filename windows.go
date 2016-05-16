package initme

import (

    "time"
    "os/exec"
    "syscall"
    "errors"

    "golang.org/x/sys/windows/svc"
    "golang.org/x/sys/windows/svc/debug"
    "log"
    "fmt"
)

type WindowsService struct {
    Name string
    Job func()
    Type string
    Start string
    Error string
    BinPath string
    Group string
    Tag string
    Depend string
    Obj string
    DisplayName string
    Password string

    Log *log.Logger

    eventLog debug.Log

}

func (self WindowsService) Register() (output string, err error, code int)  {
    args, err := self.buildScArgs("create")
    if err != nil {
        return
    }
    return self.execute(args...)
}

func (self WindowsService) Delete() (output string, err error, code int)  {
    return self.execute("delete", self.Name)
}

// https://support.microsoft.com/en-us/kb/251192
func (self WindowsService) buildScArgs(init... string) (args []string, err error) {
    args = make([]string, 0)

    args = append(args, init...)

    if self.Name != "" {
        args = append(args, self.Name)
    } else {
        return nil, errors.New("Name is mandatory")
    }
    if self.Type != "" {
        args = append(args, "type= " + self.Type)
    }
    if self.Start != "" {
        args = append(args, "start= " + self.Start)
    }
    if self.Error != "" {
        args = append(args, "error= " + self.Error)
    }
    if self.BinPath != "" {
        args = append(args, "binpath=" + self.BinPath)
    } else {
        return nil, errors.New("BinPath is mandatory")
    }
    if self.Group != "" {
        args = append(args, "group= " + self.Group)
    }
    if self.Tag != "" {
        args = append(args, "tag= " + self.Tag)
    }
    if self.Depend != "" {
        args = append(args, "depend= " + self.Depend)
    }
    if self.Obj != "" {
        if self.Password != "" {
            return nil, errors.New("Password is mandatory if Obj is set")
        }
        args = append(args, "obj= " + self.Obj)
    }
    if self.DisplayName != "" {
        args = append(args, "DisplayName= " + self.DisplayName)
    }
    if self.Password != "" {
        if self.Obj != "" {
            return nil, errors.New("Password is meanful only if Obj is set")
        }
        args = append(args, "password= " + self.Password)
    }

    return
}

func (self WindowsService) execute(args... string) (output string, err error, code int) {
    cmd := exec.Command("sc.exe", args...)
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

func (self *WindowsService) Run() {
    svc.Run(self.Name, self)
}

func (self *WindowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

    go self.Job()

loop:
	for {
		select {
		case <-tick:

			//self.eventLog.Info(1, "beep")
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				self.Log.Println(fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}