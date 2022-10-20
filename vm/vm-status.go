package vm

import (
	"context"
	"sort"
	"strings"
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

type VmSize struct {
	Vendor                string  // Clolud vendor, azure, aws, gcp, etc
	Name                  string  // name of the size
	Family                string  // Family of the size
	Size                  string  // Size ID that the vendor recoknizes
	MaxNics               int     // Max Network Interfaces
	AcceleratedNetworking bool    // Supports Accelerated Networking
	VCPU                  int     // Virtual CPUs
	PremiumIO             bool    // Supports Premium SSD / IO
	MemoryGB              float64 // Memory assigned in GB
	GpuVendor             string  // Vendor for the GPU, expect <blank>, "nvidia", "amd", "random"
	GPU                   float64 // Number of virtual GPIs
	Enabled               bool    // If the Size is enabled
	Notes                 string  // Any admin level notes
	CpuVendor             string  // Vendor of the CPU, "intel", "amd"
	CpuGeneration         string  // Generation of the CPU "Ivy Bridge", etc.
	Cost                  float64 // Cost per hour
	QuotaAvailable        int64   // Number of CPUs that are available
}

type VmSizeRequest struct {
	// accelerated networking
	AcceleratedNetworking bool `json:"AcceleratedNetworking,omitempty"`

	// CPU generation
	CPUGeneration string `json:"CPUGeneration,omitempty"`

	// CPU vendor
	CPUVendor string `json:"CPUVendor,omitempty"`

	// g p u vendor
	GPUVendor string `json:"GPUVendor,omitempty"`

	// max CPU
	MaxCPU float64 `json:"MaxCPU,omitempty"`

	// max g p u
	MaxGPU float64 `json:"MaxGPU,omitempty"`

	// max memory
	MaxMemory float64 `json:"MaxMemory,omitempty"`

	// min CPU
	MinCPU float64 `json:"MinCPU,omitempty"`

	// min g p u
	MinGPU float64 `json:"MinGPU,omitempty"`

	// min memory
	MinMemory float64 `json:"MinMemory,omitempty"`

	// name
	Name string `json:"Name,omitempty"`

	// premium i o
	PremiumIO bool `json:"PremiumIO,omitempty"`

	// specific size
	SpecificSize string `json:"SpecificSize,omitempty"`

	// vendor
	Vendor string `json:"Vendor,omitempty"`
}

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
	Size                  *VmSize           `json:"size,omitempty"`
	SizeRequest           *VmSizeRequest
	OSType                string
	OSDisk                *VirtualMachineDisk
	Disks                 []*VirtualMachineDisk
	Image                 string
	ImageVersion          string
	PrimaryNetwork        *VirtualMachineNetwork
	Networks              []*VirtualMachineNetwork
	DomainControllers     []*string
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
	Name string
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
	Delete(ctx context.Context, vm *VirtualMachineConfiguration) (*VirtualMachineConfiguration, error)
	GetLimits(ctx context.Context) ([]*VirtualMachineLimit, error)
	GetVMSizes(ctx context.Context) (map[string]*VmSize, error)
}

func FindBestVmSizes(sizeRequest *VmSizeRequest, availableSizes []*VmSize) []*VmSize {
	var toSort []*VmSize
	for _, s := range availableSizes {
		if sizeRequest.Matches(s) {
			toSort = append(toSort, s)
		}
	}

	if len(toSort) == 0 {
		return nil
	}

	sort.Sort(VMSizeCollection(toSort))
	return toSort
}

func FindBestVmSize(sizeRequest *VmSizeRequest, availableSizes []*VmSize) *VmSize {
	sorted := FindBestVmSizes(sizeRequest, availableSizes)
	if sorted == nil {
		return nil
	}
	return sorted[0]
}

func (req *VmSizeRequest) Matches(size *VmSize) bool {
	if !size.Enabled {
		return false
	}

	if req.SpecificSize != "" {
		if req.SpecificSize == size.Size {
			return true
		} else {
			return false
		}
	}

	if req.AcceleratedNetworking && !size.AcceleratedNetworking {
		return false
	}
	if req.PremiumIO && !size.PremiumIO {
		return false
	}
	if req.CPUVendor != "" && strings.EqualFold(req.CPUVendor, size.CpuVendor) {
		return false
	}
	if req.GPUVendor != "" && strings.EqualFold(req.GPUVendor, size.GpuVendor) {
		return false
	}
	if req.CPUGeneration != "" && strings.EqualFold(req.CPUGeneration, size.CpuGeneration) {
		return false
	}
	if float64(size.VCPU) < req.MinCPU {
		return false
	}
	if float64(size.VCPU) >= req.MaxCPU {
		return false
	}

	if size.GPU > 0 && req.MinGPU > 0 {
		if float64(size.GPU) < req.MinGPU {
			return false
		}
		if float64(size.GPU) >= req.MaxGPU {
			return false
		}
	}

	return true
}

type VMSizeCollection []*VmSize

func (coll VMSizeCollection) Len() int {
	return len(coll)
}

func (coll VMSizeCollection) Swap(i, j int) {
	coll[i], coll[j] = coll[j], coll[i]
}

func (coll VMSizeCollection) Less(i, j int) bool {
	s1 := coll[i]
	s2 := coll[j]

	if s1.AcceleratedNetworking && !s2.AcceleratedNetworking {
		return true
	}
	if s1.PremiumIO && !s2.PremiumIO {
		return true
	}
	if s1.MemoryGB < s2.MemoryGB {
		return true
	}
	if s1.VCPU < s2.VCPU {
		return true
	}
	if s1.GPU < s2.GPU {
		return true
	}
	if s1.QuotaAvailable > s2.QuotaAvailable {
		return true
	}

	return false
}

func FindLimit(limits []*VirtualMachineLimit, size string) *VirtualMachineLimit {
	for _, l := range limits {
		if l.Name == size {
			return l
		}
	}
	return nil
}

/*
FILTER
	Enabled
	Requires Accelerated Networking
	Requires PremiumIO
	CPU Vendor
	GPU Vendor
	CPU Generation

SCORE
	Lowest number of CPUs that match
	Lowest number of GPUs that match
	Lowest number of Memory that matches
	Accelerated Networking Preferred
	PremiumIO Preferred

*/
