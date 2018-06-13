package router

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fission/fission/crd"
)

type FunctionBackend struct {
	name          string
	weight        int64
	currentWeight int64
}

type LoadBalancer struct {
	TriggerFunctionRefMap map[string][]*FunctionBackend // trigger -> functions along with their weights
}

func makeLoadBalancer() *LoadBalancer {
	loadBalancer := &LoadBalancer{
		TriggerFunctionRefMap: make(map[string][]*FunctionBackend, 0),
	}
	return loadBalancer
}

func getCacheKey(triggerName string, triggerNamespace string, triggerResourceVersion string) string {
	return fmt.Sprintf("%v-%v-%v", triggerName, triggerNamespace, triggerResourceVersion)
}

func (lb *LoadBalancer) addFunctionBackends(trigger *crd.HTTPTrigger, functions []*FunctionBackend) {
	key := getCacheKey(trigger.Metadata.Name, trigger.Metadata.Namespace, trigger.Metadata.ResourceVersion)
	lb.TriggerFunctionRefMap[key] = functions
}

func (lb *LoadBalancer) getFunctionBackends(trigger *crd.HTTPTrigger) []*FunctionBackend {
	key := getCacheKey(trigger.Metadata.Name, trigger.Metadata.Namespace, trigger.Metadata.ResourceVersion)
	return lb.TriggerFunctionRefMap[key]
}

func (lb *LoadBalancer) deleteFunctionBackends(trigger *crd.HTTPTrigger, functions []*FunctionBackend) {
	key := getCacheKey(trigger.Metadata.Name, trigger.Metadata.Namespace, trigger.Metadata.ResourceVersion)
	lb.TriggerFunctionRefMap[key] = nil
}

func (lb *LoadBalancer) getFnBackend(trigger *crd.HTTPTrigger, functionMap map[string]functionMetadata) (*metav1.ObjectMeta, error) {
	var fnBackends []*FunctionBackend

	// it's the first time the trigger is being added to cache or trigger has been updated or router restarted.
	fnBackends = lb.getFunctionBackends(trigger)
	if len(fnBackends) == 0 {
		fnBackends = make([]*FunctionBackend, 0)
		for _, v := range functionMap {
			fnBackend := &FunctionBackend{
				name:          v.metadata.Name,
				weight:        v.weight,
				currentWeight: v.weight,
			}
			fnBackends = append(fnBackends, fnBackend)
		}
		lb.addFunctionBackends(trigger, fnBackends)
	}

	var bestBackend *FunctionBackend
	for _, backend := range fnBackends {

		backend.currentWeight += backend.weight

		if bestBackend == nil || bestBackend.currentWeight < backend.currentWeight {
			bestBackend = backend
			fmt.Printf("bestBackend: %s\n", backend.name)
		}
	}

	if bestBackend != nil {
		bestBackend.currentWeight -= 100
	}

	return functionMap[bestBackend.name].metadata, nil
}

/*
func(lb *LoadBalancer) DumpServers() {
	fmt.Printf("Printing server info\n")

	for k, v := range lb.Servers {
		fmt.Printf("Server : %s\n", k)
		fmt.Printf("Contents : %v\n", *v)
		fmt.Printf("\n\n")
	}

	fmt.Printf("***********************************\n")
}
*/
