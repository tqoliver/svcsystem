package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/shirou/gopsutil/cpu"
	"strconv"

	//"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	//"github.com/shirou/gopsutil/process"
	"log"
	//"net"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"
)

//DiskStatus holds the disk information
type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

//SysInfo holds the system info where in the that the app is running
type SysInfo struct {
	CurrentUTC          time.Time `json:"currentUTC"`
	CurrentLocalTime    time.Time `json:"currentLocalTime"`
	GolangVersion       string    `json:"golangVersion"`
	ContainerHostName   string    `json:"containerHostName"`
	HostID              string    `json:"hostID"`
	HostName            string    `json:"hostName"`
	BootTime            string    `json:"bootTime"`
	KernelVersion       string    `json:"kernelVersion"`
	Uptime              uint64    `json:"upTime"`
	UptimeDays          uint64    `json:"uptimeDays"`
	UptimeHours         uint64    `json:"uptimeHours"`
	UptimeMinutes       uint64    `json:"uptimeMinutes"`
	OperatingSystem     string    `json:"OperatingSystem"`
	Platform            string    `json:"Platform"`
	PlatformFamily      string    `json:"PlatformFamily"`
	PlatformVersion     string    `json:"PlatformVersion"`
	VirtualSystem       string    `json:"VirtualSystem"`
	VirtualRole         string    `json:"VirtualRole"`
	CPUs                int32     `json:"CPUs"`
	AllocMemory         uint64    `json:"allocatedMemory"`
	AllocMemoryMB       uint64    `json:"allocatedMemoryMB"`
	TotalAllocMemory    uint64    `json:"totalAllocatedMemory"`
	TotalAllocMemoryMB  uint64    `json:"totalAllocatedMemoryMB"`
	TotalSystemMemory   uint64    `json:"totalSystemMem"`
	TotalSystemMemoryMB uint64    `json:"totalSystemMemMB"`
	NetworkInterfaces   [20]struct {
		Name            string `json:"networkInterfaceName"`
		HardwareAddress string `json:"hardwareAddress"`
		IPAddresses     [5]struct {
			IPAddress string `json:"ipAddress"`
		} `json:"ipAddresses"`
	} `json:"networkInterfaces"`
	Disk [10]struct {
		Path string  `json:"diskPath"`
		All  float64 `json:"TotalStorage"`
		Used float64 `json:"UsedStorage"`
		Free float64 `json:"FreeStorage"`
	}
	InfoStat [10]InfoStat
}

//InfoStat is the CPU data
type InfoStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendorId"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physicalId"`
	CoreID     string   `json:"coreId"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"modelName"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cacheSize"`
	Flags      []string `json:"flags"`
	Microcode  string   `json:"microcode"`
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/v1/info/system", Systeminfo)
	r.HandleFunc("/", Index)

	log.Fatal(http.ListenAndServe(":8000", r))
}

//Index function
func Index(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	day := t.Format("2006 January _2 03:04:05PM MST")
	fmt.Fprintf(w, "<h1>dvdService is alive...kinda sorta<br>The current time is: "+day+"</h1>")
}

//Systeminfo function
func Systeminfo(w http.ResponseWriter, r *http.Request) {
	s := SystemInfo()
	fmt.Fprintf(w, s)
}

//SystemInfo will return various information about the sytem (VM or container) in which it is running
func SystemInfo() string {

	var si SysInfo
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	si.AllocMemory = m.Alloc
	si.AllocMemoryMB = btomb(m.Alloc)
	si.TotalAllocMemory = m.TotalAlloc
	si.TotalAllocMemoryMB = btomb(m.TotalAlloc)
	si.TotalSystemMemory = m.Sys
	si.TotalSystemMemoryMB = btomb(m.Sys)
	c, _ := cpu.Info()

	si.CPUs = c[0].Cores

	si.GolangVersion = runtime.Version()
	si.ContainerHostName, _ = os.Hostname()
	si.CurrentUTC = time.Now().UTC()

	si.CurrentLocalTime = time.Now().Local()

	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	v, _ := mem.VirtualMemory()
	fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

	type InfoStat struct {
		Hostname             string `json:"hostname"`
		Uptime               uint64 `json:"uptime"`
		BootTime             uint64 `json:"bootTime"`
		Procs                uint64 `json:"procs"`           // number of processes
		OS                   string `json:"os"`              // ex: freebsd, linux
		Platform             string `json:"platform"`        // ex: ubuntu, linuxmint
		PlatformFamily       string `json:"platformFamily"`  // ex: debian, rhel
		PlatformVersion      string `json:"platformVersion"` // version of the complete OS
		KernelVersion        string `json:"kernelVersion"`   // version of the OS kernel (if available)
		VirtualizationSystem string `json:"virtualizationSystem"`
		VirtualizationRole   string `json:"virtualizationRole"` // guest or host
		HostID               string `json:"hostid"`             // ex: uuid
	}

	var his *host.InfoStat
	his, _ = host.Info()

	si.Uptime = his.Uptime

	si.OperatingSystem = his.OS
	si.Platform = his.Platform
	si.PlatformFamily = his.PlatformFamily
	si.PlatformVersion = his.PlatformVersion
	si.VirtualSystem = his.VirtualizationSystem
	si.VirtualRole = his.VirtualizationRole
	si.HostID = his.HostID
	si.HostName = his.Hostname
	si.BootTime = strconv.FormatUint(his.BootTime, 10)
	si.KernelVersion = his.KernelVersion

	si.UptimeDays = si.Uptime / (60 * 60 * 24)
	si.UptimeHours = (si.Uptime - (si.UptimeDays * 60 * 60 * 24)) / (60 * 60)
	si.UptimeMinutes = ((si.Uptime - (si.UptimeDays * 60 * 60 * 24)) - (si.UptimeHours * 60 * 60)) / 60
	interfaces, err := net.Interfaces()

	if err == nil {
		for i, interfac := range interfaces {
			if interfac.Name == "" {
				continue
			}
			addrs := interfac.Addrs
			si.NetworkInterfaces[i].Name = interfac.Name
			si.NetworkInterfaces[i].HardwareAddress = string(interfac.HardwareAddr)
			for x, addr := range addrs {
				if addr.String() != "" {
					si.NetworkInterfaces[i].IPAddresses[x].IPAddress = addr.String()
				} else {
					break
				}
			}
		}
	}

	var paths [10]string
	paths[0] = "/"

	for i, path := range paths {
		disk := DiskUsage(path)
		si.Disk[i].Path = path
		si.Disk[i].All = float64(disk.All) / float64(GB)
		si.Disk[i].Used = float64(disk.Used) / float64(GB)
		si.Disk[i].Free = float64(disk.Free) / float64(GB)
	}

	strJSON, err := json.Marshal(si)
	checkErr(err)

	return string(strJSON)
}

func btomb(b uint64) uint64 {
	return b / 1024
}

// DiskUsage disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {

	if path == "" {
		return
	}

	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		log.Fatal(err)
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return disk
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
