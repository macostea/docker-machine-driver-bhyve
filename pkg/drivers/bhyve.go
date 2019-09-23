package drivers

import (
	"fmt"
	"github.com/cloudflare/cfssl/log"
	"os"
	"os/exec"
	"time"
)

type Bhyve struct {
	Name string
	Pid int
	StateDir string
	ISOImagePath string
	Grub bool
	DeviceMapPath string
	GrubCfgPath string
	DiskPath string
	process *os.Process
}

func NewBhyve(statedir string) (*Bhyve, error) {
	b := Bhyve{}
	b.StateDir = statedir

	return &b, nil
}

func (b *Bhyve) Start() (chan error, error) {
	if b.Grub {
		if err := b.createBootFiles(); err != nil {
			return nil, err
		}
		cmd, err := b.bootGrub()
		if err != nil {
			return nil, err
		}

		errCh := make(chan error, 1)
		go func() {
			log.Debugf("bhyve: Waiting for %#v", cmd)
			errCh <- cmd.Wait()
		}()

		return errCh, nil
	}

	return nil, nil
}

func (b *Bhyve) bootGrub() (*exec.Cmd, error) {
	deviceMapPath := b.DeviceMapPath
	grubCfgPath := b.GrubCfgPath

	cmd := exec.Command("/usr/local/sbin/grub-bhyve", "-m", deviceMapPath, "-g", grubCfgPath, "-M", "1024M", b.Name)
	cmd.Env = os.Environ()

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	b.Pid = cmd.Process.Pid
	b.process = cmd.Process
	log.Debugf("bhyve: Pid is %v", b.Pid)

	return cmd, nil
}

func (b *Bhyve) openTTY() *os.File {
	path := fmt.Sprintf("%s/tty", b.StateDir)
	for {
		if res, err := os.OpenFile(path, os.O_RDONLY, 0); err != nil {
			log.Infof("bhyve: openTTY: %v, retrying", err)
			time.Sleep(10 * time.Millisecond)
		} else {
			log.Infof("bhyve: openTTY: got %v", path)
			saneTerminal(res)
			setRaw(res)
			return res
		}
	}
}

func (b *Bhyve) createDeviceMap() error {
	deviceMapPath := b.DeviceMapPath
	if _, err := os.Stat(deviceMapPath); os.IsNotExist(err) {
		file, err := os.Create(deviceMapPath)
		if err != nil {
			return err
		}
		defer file.Close()

		diskPath := b.DiskPath
		resolvedIsoPath := b.ISOImagePath

		_, err = file.WriteString("(hd0) " + diskPath + "\n" +
			"(cd0) " + resolvedIsoPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bhyve) createGrubCfg() error {
	grubCfgPath := b.GrubCfgPath
	if _, err := os.Stat(grubCfgPath); os.IsNotExist(err) {
		file, err := os.Create(grubCfgPath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString("linux (cd0)/boot/vmlinuz loglevel=3 user=docker nomodeset norestore base" + "\n" +
			"initrd (cd0)/boot/initrd.img" + "\n" +
			"boot")
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bhyve) createBootFiles() error {
	if err := b.createDeviceMap(); err != nil {
		return err
	}
	if err := b.createGrubCfg(); err != nil {
		return err
	}

	return nil
}
