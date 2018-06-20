
/*
Copyright 2016 The Fission Authors.

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

package router

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fission/fission/cache"
	"github.com/fission/fission/crd"
)

type FunctionBackend struct {
	name          string
	weight        int64
	currentWeight int64
}

type LoadBalancer struct {
	FunctionBackendsForUrl *cache.Cache // trigger -> functions along with their weights
}

func makeLoadBalancer() *LoadBalancer {
	loadBalancer := &LoadBalancer{
		FunctionBackendsForUrl: cache.MakeCache(0, 0),
	}
	return loadBalancer
}

func (lb *LoadBalancer) addFunctionBackends(trigger crd.HTTPTrigger, functions []FunctionBackend) {
	mk := keyFromMetadata(&trigger.Metadata)
	err, _ := lb.FunctionBackendsForUrl.ForceSet(*mk, functions)
	if err != nil {
		// TODO: return err and check old value
		// if *serviceUrl == *(old.(*url.URL)) {
		//	return
		//}

		// ignore error
		log.Printf("error: %v caching function backends for url : %v", err, trigger.Spec.RelativeURL)
	}
}

func (lb *LoadBalancer) getFunctionBackends(trigger crd.HTTPTrigger) (fnBackendList []FunctionBackend, err error) {
	mk := keyFromMetadata(&trigger.Metadata)
	item, err := lb.FunctionBackendsForUrl.Get(*mk)
	if err != nil {
		log.Printf("Error: %v getting function backends for trigger url : %v", err, trigger.Spec.RelativeURL, )
		return
	}

	fnBackendList, ok := item.([]FunctionBackend )
	if !ok {
		log.Printf("Error typecasting item to array of FunctionBackend")
		return
	}
	return
}

func (lb *LoadBalancer) deleteFunctionBackends(trigger crd.HTTPTrigger, functions []*FunctionBackend) error{
	mk := keyFromMetadata(&trigger.Metadata)
	return lb.FunctionBackendsForUrl.Delete(*mk)
}

func (lb *LoadBalancer) getCanaryBackend(trigger *crd.HTTPTrigger, functionMap map[string]functionMetadata) (*metav1.ObjectMeta, error) {
	log.Printf("Requesting loadbalancer to choose a function backend for url : %s", trigger.Spec.RelativeURL)

	fnBackends, err := lb.getFunctionBackends(*trigger)
	if err != nil {
		log.Printf("Error getting function backends for url : %v", trigger.Spec.RelativeURL)
		return  nil, err
	}

	updatedFnBackends := make([]FunctionBackend, 0)
	bestBackend := -1

	for index, backend := range fnBackends {
		fnBackend := FunctionBackend{
			name:          backend.name,
			weight:        backend.weight,
			currentWeight: backend.currentWeight,
		}
		fnBackend.currentWeight += fnBackend.weight

		log.Printf("Just appeneded fnBackend : %+v to updatedFnBackends", fnBackend)
		updatedFnBackends = append(updatedFnBackends, fnBackend)

		if bestBackend == -1 || updatedFnBackends[bestBackend].currentWeight < fnBackend.currentWeight {
			log.Printf("setting bestBackend : %v", backend.name)
			bestBackend = index
		}

	}

	if bestBackend != -1 {
		updatedFnBackends[bestBackend].currentWeight -= 100
		log.Printf("final bestBackend's currentWeight : %+v", bestBackend)
	}

	log.Printf("updatedFnBackends : %+v", updatedFnBackends)
	lb.addFunctionBackends(*trigger, updatedFnBackends)

	log.Printf("Trying to access functionMap[%s] = %+v", updatedFnBackends[bestBackend].name, functionMap[updatedFnBackends[bestBackend].name])
	return functionMap[updatedFnBackends[bestBackend].name].metadata, nil
}
