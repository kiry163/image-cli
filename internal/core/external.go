package core

import "os/exec"

func ImageMagickCommand() (string, bool) {
	if path, err := exec.LookPath("magick"); err == nil {
		return path, true
	}
	if path, err := exec.LookPath("convert"); err == nil {
		return path, true
	}
	return "", false
}

func HasImageMagick() bool {
	_, ok := ImageMagickCommand()
	return ok
}

func ImageMagickIdentifyCommand() (string, []string, bool) {
	if path, err := exec.LookPath("magick"); err == nil {
		return path, []string{"identify"}, true
	}
	if path, err := exec.LookPath("identify"); err == nil {
		return path, nil, true
	}
	return "", nil, false
}
