/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	vmSchema "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/log"
	"kubevirt.io/kubevirt/pkg/hooks"
	hooksInfo "kubevirt.io/kubevirt/pkg/hooks/info"
	hooksV1alpha1 "kubevirt.io/kubevirt/pkg/hooks/v1alpha1"
	hooksV1alpha2 "kubevirt.io/kubevirt/pkg/hooks/v1alpha2"
)

type infoServer struct {
	Version string
}

func (s infoServer) Info(ctx context.Context, params *hooksInfo.InfoParams) (*hooksInfo.InfoResult, error) {
	log.Log.Info("Hook's Info method has been called")

	return &hooksInfo.InfoResult{
		Name: "custom",
		Versions: []string{
			s.Version,
		},
		HookPoints: []*hooksInfo.HookPoint{
			&hooksInfo.HookPoint{
				Name:     hooksInfo.OnDefineDomainHookPointName,
				Priority: 0,
			},
		},
	}, nil
}

type v1alpha1Server struct{}
type v1alpha2Server struct{}

func (s v1alpha2Server) OnDefineDomain(ctx context.Context, params *hooksV1alpha2.OnDefineDomainParams) (*hooksV1alpha2.OnDefineDomainResult, error) {
	log.Log.Info("Hook's OnDefineDomain callback method has been called")
	newDomainXML, err := onDefineDomain(params.GetVmi(), params.GetDomainXML())
	if err != nil {
		return nil, err
	}
	return &hooksV1alpha2.OnDefineDomainResult{
		DomainXML: newDomainXML,
	}, nil
}
func (s v1alpha2Server) PreCloudInitIso(_ context.Context, params *hooksV1alpha2.PreCloudInitIsoParams) (*hooksV1alpha2.PreCloudInitIsoResult, error) {
	return &hooksV1alpha2.PreCloudInitIsoResult{
		CloudInitData: params.GetCloudInitData(),
	}, nil
}

func (s v1alpha1Server) OnDefineDomain(ctx context.Context, params *hooksV1alpha1.OnDefineDomainParams) (*hooksV1alpha1.OnDefineDomainResult, error) {
	log.Log.Info("Hook's OnDefineDomain callback method has been called")
	newDomainXML, err := onDefineDomain(params.GetVmi(), params.GetDomainXML())
	if err != nil {
		return nil, err
	}
	return &hooksV1alpha1.OnDefineDomainResult{
		DomainXML: newDomainXML,
	}, nil
}

func onDefineDomain(vmiJSON []byte, domainXML []byte) ([]byte, error) {
	log.Log.Info("Hook's OnDefineDomain callback method has been called")

	vmiSpec := vmSchema.VirtualMachineInstance{}
	err := json.Unmarshal(vmiJSON, &vmiSpec)
	if err != nil {
		log.Log.Reason(err).Errorf("Failed to unmarshal given VMI spec: %s", vmiJSON)
		panic(err)
	}

	annotations := vmiSpec.GetAnnotations()

	m, merr := mxj.NewMapXml(domainXML)
	if merr != nil || m == nil {
		log.Log.Reason(merr).Errorf("Failed to unmarshal given domain spec: %s", domainXML)
		panic(merr)
	}

	for k, v := range annotations {
		if strings.HasPrefix(k, "custom.kubevirt.io/") {
			path := k[19:]
			i, err2 := ensureFinalPathExists(path, m)
			if err2 != nil {
				return i, err2
			}
			_, err := m.UpdateValuesForPath(v, path)
			if err != nil {
				log.Log.Reason(merr).Errorf("Failed to set value %s to path %s with spec: %s", v, path, domainXML)
				continue
			}
		}
	}

	newDomainXML, err := m.Xml()
	if err != nil {
		log.Log.Reason(err).Errorf("Failed to marshal updated domain spec")
		panic(err)
	}

	log.Log.Info("Successfully updated original domain spec with requested annotations")

	return newDomainXML, nil
}

func ensureFinalPathExists(path string, m mxj.Map) ([]byte, error) {
	pathSplitted := strings.Split(path, ".")
	var vp = pathSplitted[0]
	for _, p := range pathSplitted[1:] {
		if !m.Exists(vp) {
			err := m.SetValueForPath(map[string]interface{}{}, vp)
			if err != nil {
				return nil, err
			}
		}
		vp = vp + "." + p
	}
	return nil, nil
}

func main() {
	log.InitializeLogging("custom-hook-sidecar")

	var version string
	pflag.StringVar(&version, "version", "v1alpha2", "hook version to use")
	pflag.Parse()

	socketPath := hooks.HookSocketsSharedDirectory + "/customhook.sock"
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Log.Reason(err).Errorf("Failed to initialized socket on path: %s", socket)
		log.Log.Error("Check whether given directory exists and socket name is not already taken by other file")
		panic(err)
	}
	defer os.Remove(socketPath)

	server := grpc.NewServer([]grpc.ServerOption{}...)

	//hooksV1alpha1.Version,
	hooksInfo.RegisterInfoServer(server, infoServer{Version: version})
	hooksV1alpha1.RegisterCallbacksServer(server, v1alpha1Server{})
	hooksV1alpha2.RegisterCallbacksServer(server, v1alpha2Server{})
	log.Log.Infof("Starting hook server exposing 'info' and 'v1alpha1' services on socket %s", socketPath)
	server.Serve(socket)
}
