# initme
Register your Go program as service for SysV, SystemD, Upstart and Windows

You may found interesting this repository: https://github.com/kardianos/service

How to install service:
```
func (self Service) doInstall() (err error) {
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
			Description: "Plesk Slack notification service",
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
			Description: "Plesk Slack notification service",
			Exec:        self.binaryPath,
			Log:         self.log,
		}
	}
	if runtime.GOOS == "linux" && initme.IsSystemD() {
		self.log.Println("systemd")
		conf = initme.Config{
			Name:            self.Name,
			Description:     "My some test",
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
func (self Service) doUninstall() (err error) {
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
