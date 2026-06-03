//go:build !windows

package ui

func GetAudioDeviceNames() []string {
	return []string{"Predeterminado"}
}

func resolveAudioDeviceID(name string) uintptr {
	return 0
}

func startAlertAudio(deviceID uintptr, cancel <-chan struct{}) {
}
