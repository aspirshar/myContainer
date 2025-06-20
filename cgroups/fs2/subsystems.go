package fs2

import (
	"github.com/aspirshar/myContainer/cgroups/resource"
)

var Subsystems = []resource.Subsystem{
	&CpusetSubSystem{},
	&MemorySubSystem{},
	&CpuSubSystem{},
}