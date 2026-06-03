package ui

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/faiface/beep/mp3"
)

var (
	winmm               = syscall.NewLazyDLL("winmm.dll")
	waveOutGetNumDevs   = winmm.NewProc("waveOutGetNumDevs")
	waveOutGetDevCapsW  = winmm.NewProc("waveOutGetDevCapsW")
	waveOutOpen         = winmm.NewProc("waveOutOpen")
	waveOutPrepareHeader = winmm.NewProc("waveOutPrepareHeader")
	waveOutWrite        = winmm.NewProc("waveOutWrite")
	waveOutUnprepareHeader = winmm.NewProc("waveOutUnprepareHeader")
	waveOutReset        = winmm.NewProc("waveOutReset")
	waveOutClose        = winmm.NewProc("waveOutClose")
)

const (
	waveMapper    = 0xFFFFFFFF
	mmsyserrNoerr = 0
	whdrDone      = 0x00000001
	whdrPrepared  = 0x00000002
)

type waveOutCaps struct {
	wMid            uint16
	wPid            uint16
	vDriverVersion  uint32
	szPname         [32]uint16
	dwFormats       uint32
	wChannels       uint16
	wReserved       uint16
	dwSupport       uint32
}

type waveFormatEx struct {
	wFormatTag      uint16
	nChannels       uint16
	nSamplesPerSec  uint32
	nAvgBytesPerSec uint32
	nBlockAlign     uint16
	wBitsPerSample  uint16
	cbSize          uint16
}

type waveHdr struct {
	lpData         uintptr
	dwBufferLength uint32
	dwBytesRecorded uint32
	dwUser         uintptr
	dwFlags        uint32
	dwLoops        uint32
	lpNext         uintptr
	reserved       uintptr
}

func GetAudioDeviceNames() []string {
	ret, _, _ := waveOutGetNumDevs.Call()
	numDevs := int(ret)

	devices := make([]string, 0, numDevs+1)
	devices = append(devices, "Predeterminado")
	for i := 0; i < numDevs; i++ {
		var caps waveOutCaps
		ret, _, _ := waveOutGetDevCapsW.Call(
			uintptr(i),
			uintptr(unsafe.Pointer(&caps)),
			unsafe.Sizeof(caps),
		)
		if ret == mmsyserrNoerr {
			name := syscall.UTF16ToString(caps.szPname[:])
			devices = append(devices, name)
		}
	}
	return devices
}

func resolveAudioDeviceID(name string) uintptr {
	if name == "" || name == "Predeterminado" {
		return waveMapper
	}
	ret, _, _ := waveOutGetNumDevs.Call()
	numDevs := int(ret)
	for i := 0; i < numDevs; i++ {
		var caps waveOutCaps
		ret, _, _ := waveOutGetDevCapsW.Call(
			uintptr(i),
			uintptr(unsafe.Pointer(&caps)),
			unsafe.Sizeof(caps),
		)
		if ret == mmsyserrNoerr {
			devName := syscall.UTF16ToString(caps.szPname[:])
			if devName == name {
				return uintptr(i)
			}
		}
	}
	return waveMapper
}

type pcmAudio struct {
	data        []byte
	sampleRate  int
	numChannels int
}

func decodeAlertToPCM() (*pcmAudio, error) {
	rutaAudio := filepath.Join("assets", "sounds", "alerta1.mp3")
	f, err := os.Open(rutaAudio)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()

	numCh := format.NumChannels
	if numCh == 0 {
		numCh = 2
	}
	var samples []float64
	buf := make([][2]float64, 2048)
	for {
		n, ok := streamer.Stream(buf)
		if !ok {
			break
		}
		for i := 0; i < n; i++ {
			for ch := 0; ch < numCh; ch++ {
				samples = append(samples, buf[i][ch])
			}
		}
	}

	pcm := make([]byte, len(samples)*2)
	for i, s := range samples {
		val := int16(clampFloat(s) * math.MaxInt16)
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(val))
	}

	return &pcmAudio{
		data:        pcm,
		sampleRate:  int(format.SampleRate),
		numChannels: numCh,
	}, nil
}

func clampFloat(v float64) float64 {
	if v > 1.0 {
		return 1.0
	}
	if v < -1.0 {
		return -1.0
	}
	return v
}

func startAlertAudio(deviceID uintptr, cancel <-chan struct{}) {
	pcm, err := decodeAlertToPCM()
	if err != nil {
		return
	}

	blockAlign := uint16(pcm.numChannels * 2)
	wfx := waveFormatEx{
		wFormatTag:      1,
		nChannels:       uint16(pcm.numChannels),
		nSamplesPerSec:  uint32(pcm.sampleRate),
		nAvgBytesPerSec: uint32(pcm.sampleRate) * uint32(blockAlign),
		nBlockAlign:     blockAlign,
		wBitsPerSample:  16,
	}

	for {
		select {
		case <-cancel:
			return
		default:
		}

		var hWave uintptr
		ret, _, _ := waveOutOpen.Call(
			uintptr(unsafe.Pointer(&hWave)),
			deviceID,
			uintptr(unsafe.Pointer(&wfx)),
			0, 0, 0,
		)
		if ret != mmsyserrNoerr {
			return
		}

		hdr := waveHdr{
			lpData:         uintptr(unsafe.Pointer(&pcm.data[0])),
			dwBufferLength: uint32(len(pcm.data)),
		}

		ret, _, _ = waveOutPrepareHeader.Call(hWave, uintptr(unsafe.Pointer(&hdr)), unsafe.Sizeof(hdr))
		if ret != mmsyserrNoerr {
			waveOutClose.Call(hWave)
			return
		}

		ret, _, _ = waveOutWrite.Call(hWave, uintptr(unsafe.Pointer(&hdr)), unsafe.Sizeof(hdr))
		if ret != mmsyserrNoerr {
			waveOutUnprepareHeader.Call(hWave, uintptr(unsafe.Pointer(&hdr)), unsafe.Sizeof(hdr))
			waveOutClose.Call(hWave)
			return
		}

		done := false
		for !done {
			select {
			case <-cancel:
				waveOutReset.Call(hWave)
				done = true
			default:
				if hdr.dwFlags&whdrDone != 0 {
					done = true
				} else {
					time.Sleep(30 * time.Millisecond)
				}
			}
		}

		waveOutUnprepareHeader.Call(hWave, uintptr(unsafe.Pointer(&hdr)), unsafe.Sizeof(hdr))
		waveOutClose.Call(hWave)
	}
}
