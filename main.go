package main

import (
	"vcluster-nodeport-plugin/hook"

	"os"
	"strconv"
	"strings"

	"github.com/loft-sh/vcluster-sdk/plugin"
	log "github.com/sirupsen/logrus"
)

func main() {
	_ = plugin.MustInit()

	portsStr := os.Getenv("PORT_MAPPINGS")
	labelSelectorStr := os.Getenv("LABEL_SELECTOR")

	ports := ParsePortMappings(portsStr)
	labelSelector := ParseLabelSelector(labelSelectorStr)
	serviceHook := hook.NewServiceHook(ports, labelSelector)
	plugin.MustRegister(serviceHook)
	plugin.MustStart()
}

func ParsePortMappings(portsStr string) map[string]int32 {
	portMap := make(map[string]int32)
	mappings := strings.Split(portsStr, ",")
	for _, mapping := range mappings {
		parts := strings.Split(strings.TrimSpace(mapping), ":")
		if len(parts) == 2 {
			portName := strings.TrimSpace(parts[0])
			portValue, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				log.Warnf("Invalid port value for %s: %s", portName, parts[1])
				continue
			}
			portMap[portName] = int32(portValue)
			log.Infof("Added port mapping: %s -> %d", portName, portValue)
		}

	}

	return portMap
}

func ParseLabelSelector(selectorStr string) map[string]string {
	labelMap := make(map[string]string)
	if selectorStr == "" {
		return labelMap
	}

	selectors := strings.Split(selectorStr, ",")
	for _, selector := range selectors {
		parts := strings.Split(strings.TrimSpace(selector), "=")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			labelMap[key] = value
			log.Infof("Added label selector: %s=%s", key, value)
		}
	}

	return labelMap
}
