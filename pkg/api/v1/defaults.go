package v1

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/pborman/uuid"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

var _true = t(true)
var _false = t(false)

func SetDefaults_HPETTimer(obj *HPETTimer) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_PITTimer(obj *PITTimer) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_KVMTimer(obj *KVMTimer) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_HypervTimer(obj *HypervTimer) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_RTCTimer(obj *RTCTimer) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_FeatureState(obj *FeatureState) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_FeatureAPIC(obj *FeatureAPIC) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_FeatureVendorID(obj *FeatureVendorID) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
}

func SetDefaults_DiskDevice(obj *DiskDevice) {
	if obj.Disk == nil &&
		obj.CDRom == nil &&
		obj.Floppy == nil &&
		obj.LUN == nil {
		obj.Disk = &DiskTarget{}
	}
}

func SetDefaults_Watchdog(obj *Watchdog) {
	if obj.I6300ESB == nil {
		obj.I6300ESB = &I6300ESBWatchdog{}
	}
}

func SetDefaults_CDRomTarget(obj *CDRomTarget) {
	if obj.ReadOnly == nil {
		obj.ReadOnly = _true
	}
	if obj.Tray == "" {
		obj.Tray = TrayStateClosed
	}
}

func SetDefaults_FloppyTarget(obj *FloppyTarget) {
	if obj.Tray == "" {
		obj.Tray = TrayStateClosed
	}
}

func SetDefaults_FeatureSpinlocks(obj *FeatureSpinlocks) {
	if obj.Enabled == nil {
		obj.Enabled = _true
	}
	if *obj.Enabled == *_true && obj.Retries == nil {
		obj.Retries = ui32(4096)
	}
}

func SetDefaults_I6300ESBWatchdog(obj *I6300ESBWatchdog) {
	if obj.Action == "" {
		obj.Action = WatchdogActionReset
	}
}

func SetDefaults_Firmware(obj *Firmware) {
	if obj.UUID == "" {
		obj.UUID = types.UID(uuid.NewRandom().String())
	}
}

func SetDefaults_VirtualMachine(obj *VirtualMachine) {
	// FIXME we need proper validation and configurable defaulting instead of this
	if _, exists := obj.Spec.Domain.Resources.Requests[v1.ResourceMemory]; !exists {
		obj.Spec.Domain.Resources.Requests = v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("8192Ki"),
		}
	}
	if obj.Spec.Domain.Firmware == nil {
		obj.Spec.Domain.Firmware = &Firmware{}
	}

	if obj.Spec.Domain.Features == nil {
		obj.Spec.Domain.Features = &Features{}
	}
	if obj.Spec.Domain.Machine.Type == "" {
		obj.Spec.Domain.Machine.Type = "q35"
	}
	setDefaults_DiskFromMachineType(obj)
	setDefaults_NetworkInterface(obj)
}

func setDefaults_DiskFromMachineType(obj *VirtualMachine) {
	bus := diskBusFromMachine(obj.Spec.Domain.Machine.Type)

	for i := range obj.Spec.Domain.Devices.Disks {
		disk := &obj.Spec.Domain.Devices.Disks[i].DiskDevice

		SetDefaults_DiskDevice(disk)

		if disk.Disk != nil && disk.Disk.Bus == "" {
			disk.Disk.Bus = bus
		}
		if disk.CDRom != nil && disk.CDRom.Bus == "" {
			disk.CDRom.Bus = bus
		}
		if disk.LUN != nil && disk.LUN.Bus == "" {
			disk.LUN.Bus = bus
		}
	}
}

func GetNumberOfPodInterfaces(spec *VirtualMachineSpec) int {
	nPodInterfaces := 0
	for _, net := range spec.Networks {
		if net.Pod != nil {
			for _, iface := range spec.Domain.Devices.Interfaces {
				if iface.Name == net.Name {
					nPodInterfaces++
					break // we maintain 1-to-1 relationship between networks and interfaces
				}
			}
		}
	}
	return nPodInterfaces
}

func setDefaults_NetworkInterface(obj *VirtualMachine) {
	if GetNumberOfPodInterfaces(&obj.Spec) == 0 {
		iface, net := GetDefaultInterfaceAndNetwork()
		obj.Spec.Domain.Devices.Interfaces = append(obj.Spec.Domain.Devices.Interfaces, *iface)
		obj.Spec.Networks = append(obj.Spec.Networks, *net)
	}
}

func getRandomString(length int) string {
	runes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, length)
	for i := range s {
		s[i] = runes[rand.Intn(len(runes))]
	}
	return string(s)
}

func GetDefaultInterfaceAndNetwork() (*Interface, *Network) {
	iface := DefaultNetworkInterface()
	net := DefaultPodNetwork()

	postfixLength := 8
	postfix := getRandomString(postfixLength)
	name := fmt.Sprintf("default-%s", postfix)
	iface.Name = name
	net.Name = name

	return iface, net
}

func DefaultNetworkInterface() *Interface {
	// TODO:(ihar) switch consumers to GetDefaultInterfaceAndNetwork
	iface := &Interface{
		Name: "default",
		InterfaceBindingMethod: InterfaceBindingMethod{
			Bridge: &InterfaceBridge{},
		},
	}
	return iface
}

func DefaultPodNetwork() *Network {
	// TODO:(ihar) switch consumers to GetDefaultInterfaceAndNetwork
	defaultNet := &Network{
		Name: "default",
		NetworkSource: NetworkSource{
			Pod: &PodNetwork{},
		},
	}
	return defaultNet
}

func diskBusFromMachine(machine string) string {
	// catches: "q35", "pc-q35-*"
	// see /path/to/qemu-kvm -machine help
	if strings.HasPrefix(machine, "pc-q35") || strings.HasPrefix(machine, "q35") {
		return "sata"
	}
	// safe fallback for x86_64, but very slow
	return "ide"
}

func t(v bool) *bool {
	return &v
}

func ui32(v uint32) *uint32 {
	return &v
}
