package tray

import (
	"fyne.io/systray"
)

// TrayApp represents the system tray application
type TrayApp struct {
	onShow         func()
	onDiskAnalyzer func()
	onSettings     func()
	onQuit         func()
}

// New creates a new TrayApp instance
func New() *TrayApp {
	return &TrayApp{}
}

// SetCallbacks sets the callback functions for tray menu items
func (t *TrayApp) SetCallbacks(onShow, onDiskAnalyzer, onSettings, onQuit func()) {
	t.onShow = onShow
	t.onDiskAnalyzer = onDiskAnalyzer
	t.onSettings = onSettings
	t.onQuit = onQuit
}

// Run starts the system tray
func (t *TrayApp) Run() {
	systray.Run(t.onReady, t.onExit)
}

// Quit exits the system tray
func (t *TrayApp) Quit() {
	systray.Quit()
}

func (t *TrayApp) onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("ClearC")
	systray.SetTooltip("ClearC - 系统清理工具")

	// Menu items
	mShow := systray.AddMenuItem("打开主窗口", "显示 ClearC 主窗口")
	systray.AddSeparator()
	mDiskAnalyzer := systray.AddMenuItem("磁盘分析", "分析磁盘空间占用")
	systray.AddSeparator()
	mSettings := systray.AddMenuItem("设置", "打开设置页面")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "退出 ClearC")

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				if t.onShow != nil {
					t.onShow()
				}
			case <-mDiskAnalyzer.ClickedCh:
				if t.onDiskAnalyzer != nil {
					t.onDiskAnalyzer()
				}
			case <-mSettings.ClickedCh:
				if t.onSettings != nil {
					t.onSettings()
				}
			case <-mQuit.ClickedCh:
				if t.onQuit != nil {
					t.onQuit()
				}
				systray.Quit()
				return
			}
		}
	}()
}

func (t *TrayApp) onExit() {
	// Cleanup if needed
}

// getIcon returns the tray icon bytes
func getIcon() []byte {
	return GetIconData()
}
