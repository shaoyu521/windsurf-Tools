package main

func (a *App) supportsTray() bool {
	if a != nil && a.traySupportedFn != nil {
		return a.traySupportedFn()
	}
	return traySupported()
}
