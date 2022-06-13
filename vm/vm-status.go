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
	ID                    string             `json:"id,omitempty"`
	LongID                string             `json:"longId,omitempty"`
	Name                  string             `json:"name,omitempty"`
	Tags                  map[string]*string `json:"tags,omitempty"`
	Size                  string             `json:"size,omitempty"`
	OSDisk                *VirtualMachineDisk
	Disks                 []*VirtualMachineDisk
	Image                 string
	Networks              []*VirtualMachineNetwork
	Credientials          Credientials
	StartUpCommand        string
	CloudProviderSettings interface{} // Custom settings for this cloud provider
}

type Credientials struct {
	AdminUser     string
	AdminPassword string
}

type VirtualMachineNetwork struct {
}

type VirtualMachineDisk struct {
}

type VirtualMachineAction string

const (
	VirtualMachineStart     VirtualMachineAction = "start"
	VirtualMachineStop      VirtualMachineAction = "stop"
	VirtualMachineTerminate VirtualMachineAction = "terminate"
)

type VMState string

const Start = VMState("start")
const Stop = VMState("stop")
const Terminate = VMState("terminate")

type VMController interface {
	ListAll(ctx context.Context) ([]*VirtualMachineStatus, error)
	ListWithTag(ctx context.Context, tag string) ([]*VirtualMachineStatus, error)
	Status(ctx context.Context, vmName string) (*VirtualMachineStatus, error)
	SetState(ctx context.Context, state VMState, vmName string, wait bool)
	Start(ctx context.Context, vmName string, wait bool) error
	Stop(ctx context.Context, vmName string, wait bool) error
	Terminate(ctx context.Context, vmName string, wait bool) error
	Get(ctx context.Context, vmName string)
}
