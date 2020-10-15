/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/docopt/docopt-go"
	topology "sigs.k8s.io/node-feature-discovery/pkg/nfd-topology-updater"
  // topologypb "sigs.k8s.io/node-feature-discovery/pkg/topologyupdater"
	v1alpha1 "github.com/swatisehgal/topologyapi/pkg/apis/topology/v1alpha1"
	"sigs.k8s.io/node-feature-discovery/pkg/version"
	// "sigs.k8s.io/node-feature-discovery/pkg/exporter"
	"sigs.k8s.io/node-feature-discovery/pkg/finder"
	"sigs.k8s.io/node-feature-discovery/pkg/kubeconf"
	"sigs.k8s.io/node-feature-discovery/pkg/podres"
)

const (
	// ProgramName is the canonical name of this program
	ProgramName = "nfd-topology-updater"
)

func main() {
	// Assert that the version is known
	if version.Undefined() {
		log.Printf("WARNING: version not set! Set -ldflags \"-X sigs.k8s.io/node-feature-discovery/pkg/version.version=`git describe --tags --dirty --always`\" during build or run.")
	}

	// Parse command-line arguments.
	// _,finderArgs, err := argsParse(nil)
	args,finderArgs, err := argsParse(nil)
	if err != nil {
		log.Fatalf("failed to parse command line: %v", err)
	}

	klConfig, err := kubeconf.GetKubeletConfigFromLocalFile(finderArgs.KubeletConfigFile)
	if err != nil {
		log.Fatalf("error getting topology Manager Policy: %v", err)
	}
	tmPolicy := klConfig.TopologyManagerPolicy
	log.Printf("Detected kubelet Topology Manager policy %q", tmPolicy)

	podResClient, err := podres.GetPodResClient(finderArgs.PodResourceSocketPath)
	if err != nil {
		log.Fatalf("Failed to get PodResource Client: %v", err)
	}
	var finderInstance finder.Finder

	finderInstance, err = finder.NewPodResourceFinder(finderArgs, podResClient)
	if err != nil {
		log.Fatalf("Failed to initialize Finder instance: %v", err)
	}
	// crdExporter, err := exporter.NewExporter(tmPolicy)
	// if err != nil {
	// 	log.Fatalf("Failed to initialize crdExporter instance: %v", err)
	// }

	// CAUTION: these resources are expected to change rarely - if ever.
	//So we are intentionally do this once during the process lifecycle.
	//TODO: Obtain node resources dynamically from the podresource API
	zonesChannel := make(chan map[string]*v1alpha1.Zone)
  var zones	map[string]*v1alpha1.Zone
	nodeResourceData, err := finder.NewNodeResources(finderArgs.SysfsRoot, podResClient)
	if err != nil {
		log.Fatalf("Failed to obtain node resource information: %v", err)
	}
	log.Printf("nodeResourceData is: %v\n", nodeResourceData)
	go func() {
		for {
			log.Printf("Scanning\n")
			podResources, err := finderInstance.Scan(nodeResourceData.GetDeviceResourceMap())
			log.Printf("podResources is: %v\n", podResources)
			if err != nil {
				log.Printf("Scan failed: %v\n", err)
				continue
			}
			log.Printf("Aggregating\n")
			zones = finder.Aggregate(podResources, nodeResourceData)
			zonesChannel <- zones
			log.Printf("zones:%v", spew.Sdump(zones))
			// if err = crdExporter.CreateOrUpdate("default", zones); err != nil {
			// 	log.Fatalf("ERROR: %v", err)
			// }
			time.Sleep(finderArgs.SleepInterval)
		}
	}()



	log.Printf("Creating NewTopologyUpdater\n")
	// Get new TopologyUpdater instance
	instance, err := topology.NewTopologyUpdater(args, tmPolicy)
	if err != nil {
		log.Fatalf("Failed to initialize NfdWorker instance: %v", err)
	}
		for{
			log.Printf("Received value on ZoneChannel\n")
			zonesValue := <-zonesChannel
			log.Printf("Updating\n")
			if err = instance.Update(zonesValue); err != nil {
				log.Fatalf("ERROR: %v", err)
			}
			if args.Oneshot {
				break
			}
		}
}

// argsParse parses the command line arguments passed to the program.
// The argument argv is passed only for testing purposes.
func argsParse(argv []string) (topology.Args,finder.Args, error) {
	args := topology.Args{}
	finderArgs := finder.Args{}
	usage := fmt.Sprintf(`%s.

  Usage:
  %s [--no-publish] [--oneshot | --sleep-interval=<seconds>] [--server=<server>]
	   [--server-name-override=<name>] [--ca-file=<path>] [--cert-file=<path>]
		 [--key-file=<path>] [--container-runtime=<runtime>] [--podresources-socket=<path>]
		 [--watch-namespace=<namespace>] [--sysfs=<mountpoint>] [--kubelet-config-file=<path>]

  %s -h | --help
  %s --version

  Options:
  -h --help                   Show this screen.
  --version                   Output version and exit.
  --ca-file=<path>            Root certificate for verifying connections
                              [Default: ]
  --cert-file=<path>          Certificate used for authenticating connections
                              [Default: ]
  --key-file=<path>           Private key matching --cert-file
                              [Default: ]
  --server=<server>           NFD server address to connecto to.
                              [Default: localhost:8080]
  --server-name-override=<name> Name (CN) expect from server certificate, useful
                              in testing
                              [Default: ]
  --no-publish                Do not publish discovered features to the
                              cluster-local Kubernetes API server.
  --oneshot                   Label once and exit.
  --sleep-interval=<seconds>  Time to sleep between re-labeling. Non-positive
                              value implies no re-labeling (i.e. infinite
                              sleep). [Default: 60s]
  --watch-namespace=<namespace> Namespace to watch pods for. Use "" for all namespaces.
  --sysfs=<mountpoint>            Mount point of the sysfs. [Default: /host-sys]
  --kubelet-config-file=<path>    Kubelet config file path. [Default: /host-etc/kubernetes/kubelet.conf]
	--podresources-socket=<path>    Pod Resource Socket path to use. `,

		ProgramName,
		ProgramName,
		ProgramName,
		ProgramName,
	)

	arguments, _ := docopt.ParseArgs(usage, argv,
		fmt.Sprintf("%s %s", ProgramName, version.Get()))

	// Parse argument values as usable types.
	var err error
	args.CaFile = arguments["--ca-file"].(string)
	args.CertFile = arguments["--cert-file"].(string)
	args.KeyFile = arguments["--key-file"].(string)
	args.NoPublish = arguments["--no-publish"].(bool)
	args.Server = arguments["--server"].(string)
	args.ServerNameOverride = arguments["--server-name-override"].(string)
	args.Oneshot = arguments["--oneshot"].(bool)
	finderArgs.SleepInterval, err = time.ParseDuration(arguments["--sleep-interval"].(string))
	if err != nil {
		return args, finderArgs, fmt.Errorf("invalid --sleep-interval specified: %s", err.Error())
	}
	if ns, ok := arguments["--watch-namespace"].(string); ok {
		finderArgs.Namespace = ns
	}
	if kubeletConfigPath, ok := arguments["--kubelet-config-file"].(string); ok {
		finderArgs.KubeletConfigFile = kubeletConfigPath
	}
	finderArgs.SysfsRoot = arguments["--sysfs"].(string)
	if path, ok := arguments["--podresources-socket"].(string); ok {
		finderArgs.PodResourceSocketPath = path
	}

	return args, finderArgs, nil
}
