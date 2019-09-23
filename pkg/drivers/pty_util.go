package drivers

import (
	"golang.org/x/sys/unix"
	"os"
	"unsafe"
)

func tcget(fd uintptr, p *unix.Termios) error {
	return ioctl(fd, unix.TIOCGETA, uintptr(unsafe.Pointer(p)))
}

func tcset(fd uintptr, p *unix.Termios) error {
	return ioctl(fd, unix.TIOCSETA, uintptr(unsafe.Pointer(p)))
}

func ioctl(fd, flag, data uintptr) error {
	if _, _, err := unix.Syscall(unix.SYS_IOCTL, fd, flag, data); err != 0 {
		return err
	}
	return nil
}

func saneTerminal(f *os.File) error {
	var termios unix.Termios
	if err := tcget(f.Fd(), &termios); err != nil {
		return err
	}

	termios.Oflag &^= unix.ONLCR
	return tcset(f.Fd(), &termios)
}

func setRaw(f *os.File) error {
	var termios unix.Termios
	if err := tcget(f.Fd(), &termios); err != nil {
		return err
	}
	termios = cfmakeraw(termios)
	termios.Oflag = termios.Oflag | unix.OPOST
	return tcset(f.Fd(), &termios)
}

// isTerminal checks if the provided file is a terminal
func isTerminal(f *os.File) bool {
	var termios unix.Termios
	if tcget(f.Fd(), &termios) != nil {
		return false
	}
	return true
}

func cfmakeraw(t unix.Termios) unix.Termios {
	t.Iflag = t.Iflag & ^uint32(unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON)
	t.Oflag = t.Oflag & ^uint32(unix.OPOST)
	t.Lflag = t.Lflag & ^(uint32(unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN))
	t.Cflag = t.Cflag & ^(uint32(unix.CSIZE | unix.PARENB))
	t.Cflag = t.Cflag | unix.CS8
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0

	return t
}
