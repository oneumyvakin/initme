# initme
Register your Go program as service for SysV, SystemD, Upstart and Windows

You may found interesting this repository: https://github.com/kardianos/service

```
// SomeService here it's your own struct

func (self SomeService) Start() (err error) {
	// all work doing here
}
```

How to install service:
```
func (self SomeService) doInstall() (err error) {
	self.log.Println("Do Install on")

	var conf initme.Config

	if runtime.GOOS == "windows" {
		self.log.Println("Install windows")

		binPath := "\"" + self.binaryPath + "\" --config \"" + self.configPath + "\""

		conf = initme.Config{
			Name:      self.Name,
			BinPath:   binPath,
			StartType: "auto",
			Log:       self.log,
		}

	}
	if runtime.GOOS == "linux" && initme.IsSysV() {
		self.log.Println("sysv")
		conf = initme.Config{
			Name:        self.Name,
			Description: self.Description,
			Provides:    self.Name,
			Required:    "$local_fs $remote_fs $network $syslog",
			Command:     self.binaryPath,
			Log:         self.log,
		}

	}
	if runtime.GOOS == "linux" && initme.IsUpstart() {
		self.log.Println("upstart")
		conf = initme.Config{
			Name:        self.Name,
			Description: self.Description,
			Exec:        self.binaryPath,
			Log:         self.log,
		}
	}
	if runtime.GOOS == "linux" && initme.IsSystemD() {
		self.log.Println("systemd")
		conf = initme.Config{
			Name:            self.Name,
			Description:     self.Description,
			TimeoutStartSec: "1",
			ExecStart:       self.binaryPath,
			WantedBy:        "multi-user.target",
			Log:             self.log,
		}
	}

	s := initme.New(conf)
	_, err, _ = s.Register()

	return
}
```

How to uninstall service:

```
func (self SomeService) doUninstall() (err error) {
	self.log.Println("Do Uninstall on")

	conf := initme.Config{
		Name: self.Name,
		Log:  self.log,
	}

	s := initme.New(conf)

	_, err, _ = s.Stop()
	_, err, _ = s.Disable()
	_, err, _ = s.Delete()

	return
}
```

Your main():
```
func main() {
	someService := SomeService{
		Name: "MyService",
	}
	isInstall, isUninstall, config := getOpts()
	someService.SetConfig(config)

	if isInstall {
		someService.doInstall()
		return
	}
	if isUninstall {
		someService.doUninstall()
		return
	}

	if runtime.GOOS == "windows" { // this case for Windows non-interactive(as service) mode
		conf := initme.Config{
			Name:    someService.Name,
			BinPath: "\"" + someService.binaryPath + "\" --config \"" + someService.configPath + "\"",
			Job:     someService.Start, // all work doing here
			Log:     someService.log,
		}

		s := initme.New(conf)

		isIntSess, err := s.IsAnInteractiveSession()
		if err != nil {
			log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
		}

		if !isIntSess {

			s.Run()

			return
		}

	}

	someService.Start() // this case for Linux and Windows interactive(!) mode, all work doing here
}

```
