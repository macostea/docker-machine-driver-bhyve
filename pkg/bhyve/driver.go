package bhyve

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"syscall"

	pkgdrivers "github.com/macostea/docker-machine-driver-bhyve/pkg/drivers"
)

const (
	isoFilename = "boot2docker.iso"
	isoMountPath = "b2d-image"
	permErr         = "%s needs to run with elevated permissions. " +
		"Please run the following command, then try again: " +
		"sudo chown root:wheel %s && sudo chmod u+s %s"
)

type Driver struct {
	*drivers.BaseDriver
	Boot2DockerURL string
	DiskSize       int
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver {
		BaseDriver: &drivers.BaseDriver{
			SSHUser:        "docker",
			StorePath:		storePath,
		},
		DiskSize: 16384,
	}
}

func (d *Driver) PreCreateCheck() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if syscall.Geteuid() != 0 {
		return fmt.Errorf(permErr, filepath.Base(exe), exe, exe)
	}

	return nil
}

func (d *Driver) Create() error {
	if err := pkgdrivers.MakeDiskImage(d.BaseDriver, d.Boot2DockerURL, d.DiskSize); err != nil {
		return errors.Wrap(err, "making disk image")
	}

	return d.Start()
}

func (d *Driver) Start() error {
	b, err := pkgdrivers.NewBhyve(filepath.Join(d.StorePath, "machines", d.MachineName))
	if err != nil {
		return err
	}

	b.Name = "boot2docker"
	b.Grub = true
	b.GrubCfgPath = pkgdrivers.GetGrubCfgPath(d.BaseDriver)
	b.DeviceMapPath = pkgdrivers.GetDeviceMapPath(d.BaseDriver)
	b.DiskPath = pkgdrivers.GetDiskPath(d.BaseDriver)
	b.ISOImagePath = d.ResolveStorePath(isoFilename)

	if _, err := b.Start(); err != nil {
		return err
	}

	return nil
}
