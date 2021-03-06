package vm

import (
	"context"
	"time"

	"github.com/appliedres/cloudy"
)

/*

Azure:
- resource group
- imagePublisher = "MicrosoftWindowsDesktop"
- imageOffer     = "Windows-10"
- imageSKU       = "21h1-ent"
- imageVersion   = "latest"
- location
*/

var VmControllers = cloudy.NewProviderRegistry[VMController]()

type VirtualMachineStatus struct {
	ID                string             `json:"id,omitempty"`
	LongID            string             `json:"longId,omitempty"`
	Name              string             `json:"name,omitempty"`
	Tags              map[string]*string `json:"tags,omitempty"`
	User              string             `json:"user,omitempty"`
	Size              string             `json:"size,omitempty"`
	PowerState        string             `json:"powerState,omitempty"`
	ProvisioningState string             `json:"provisioningState,omitempty"`
	ProvisioningTime  time.Time          `json:"provisioningTime,omitempty"`
	OperatingSystem   string             `json:"operatingSystem,omitempty"`
}

type VirtualMachineConfiguration struct {
	ID                    string            `json:"id,omitempty"`
	LongID                string            `json:"longId,omitempty"`
	Name                  string            `json:"name,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
	Size                  string            `json:"size,omitempty"`
	OSType                string
	OSDisk                *VirtualMachineDisk
	Disks                 []*VirtualMachineDisk
	Image                 string
	ImageVersion          string
	PrimaryNetwork        *VirtualMachineNetwork
	Networks              []*VirtualMachineNetwork
	Credientials          Credientials
	StartUpCommand        string
	CloudProviderSettings interface{} // Custom settings for this cloud provider
}

type VirtualMachineLimit struct {
	Name    string
	Current int
	Limit   int
}

type Credientials struct {
	AdminUser     string
	AdminPassword string
	SSHKey        string
}

type VirtualMachineNetwork struct {
	ID        string
	Name      string
	PrivateIP string
	PublicIP  string
}

type VirtualMachineDisk struct {
	Size string
}

type VirtualMachineAction string

const (
	VirtualMachineStart     VirtualMachineAction = "start"
	VirtualMachineStop      VirtualMachineAction = "stop"
	VirtualMachineTerminate VirtualMachineAction = "terminate"
)

type VMController interface {
	ListAll(ctx context.Context) ([]*VirtualMachineStatus, error)
	ListWithTag(ctx context.Context, tag string) ([]*VirtualMachineStatus, error)
	Status(ctx context.Context, vmName string) (*VirtualMachineStatus, error)
	SetState(ctx context.Context, state VirtualMachineAction, vmName string, wait bool) (*VirtualMachineStatus, error)
	Start(ctx context.Context, vmName string, wait bool) error
	Stop(ctx context.Context, vmName string, wait bool) error
	Terminate(ctx context.Context, vmName string, wait bool) error
	Create(ctx context.Context, vm *VirtualMachineConfiguration) (*VirtualMachineConfiguration, error)
	GetLimits(ctx context.Context) ([]*VirtualMachineLimit, error)
}
