
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
	"fmt"
	"github.com/fission/fission/crd"
)
/*
func (fmap *functionServiceMap) lookup(f *metav1.ObjectMeta) (*url.URL, error) {
	mk := keyFromMetadata(f)
	item, err := fmap.cache.Get(*mk)
	if err != nil {
		return nil, err
	}
	u := item.(*url.URL)
	return u, nil
}

func (fmap *functionServiceMap) assign(f *metav1.ObjectMeta, serviceUrl *url.URL) {
	mk := keyFromMetadata(f)
	err, old := fmap.cache.Set(*mk, serviceUrl)
	if err != nil {
		log.Printf("Comparing serviceUrl with oldUrl, serviceUrl.Host = %v , oldUrl.Host = %v", serviceUrl.Host, old.(*url.URL).Host)
		log.Printf("Also dumping serviceUrl obj %+v, old : %+v", *serviceUrl, *(old.(*url.URL)))
		if *serviceUrl == *(old.(*url.URL)) {
			return
		}
		log.Printf("error caching service url for function with a different value: %v", err)
		// ignore error
	}
}

func (fmap *functionServiceMap) remove(f *metav1.ObjectMeta) error {
	mk := keyFromMetadata(f)
	return fmap.cache.Delete(*mk)
}
*/



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
		FunctionBackendsForUrl: cache.MakeCache(expiry, 0),
	}
	return loadBalancer
}

func getCacheKey(triggerName string, triggerNamespace string, triggerResourceVersion string) string {
	return fmt.Sprintf("%v-%v-%v", triggerName, triggerNamespace, triggerResourceVersion)
}

func (lb *LoadBalancer) addFunctionBackends(trigger crd.HTTPTrigger, functions []FunctionBackend) {
	mk := keyFromMetadata(&trigger.Metadata)
	err, _ := lb.FunctionBackendsForUrl.Set(*mk, functions)
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
		return fnBackendList, err
	}

	fnaBackendList, ok := item.([]FunctionBackend )
	if !ok {
		log.Printf("Error typecasting item to array of FunctionBackend")
	}
	return fnaBackendList, nil
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
	var bestBackend *FunctionBackend

	for _, backend := range fnBackends {

		backend.currentWeight += backend.weight

		if bestBackend == nil || bestBackend.currentWeight < backend.currentWeight {
			bestBackend = &backend
		}

		fnBackend := FunctionBackend{
			name:          backend.name,
			weight:        backend.weight,
			currentWeight: backend.weight,
		}

		updatedFnBackends = append(updatedFnBackends, fnBackend)
	}

	if bestBackend != nil {
		bestBackend.currentWeight -= 100
	}

	lb.addFunctionBackends(*trigger, updatedFnBackends)

	log.Printf("Trying to access functionMap[%s] = %+v", bestBackend.name, functionMap[bestBackend.name])
	return functionMap[bestBackend.name].metadata, nil
}
