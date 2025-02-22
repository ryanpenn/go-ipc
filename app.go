package main

type App struct {
	Data    map[string]any // 封装一下，同步数据时禁止改变数据
	Port    int
	Service *SyncService
}

func NewApp(port int) *App {
	a := &App{
		Data: make(map[string]any),
		Port: port,
	}
	a.Service = NewSyncService(a)
	return a
}

func (a *App) Run() error {
	return a.Service.start()
}

func (a *App) Stop() {
	a.Service.Stop()
}
