
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
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setupCanaryLoadBalancer() {
	// just seeding the random number
	rand.Seed(time.Now().UnixNano())
}

func findCeil(randomNumber int, wtDistrList []FunctionWeightDistribution) string{
	low := wtDistrList[0].sumPrefix
	high := wtDistrList[len(wtDistrList)].sumPrefix
	for {
		mid := low + high / 2
		if randomNumber > wtDistrList[mid].sumPrefix {
			high = mid
		} else {
			low = mid + 1
		}

		if low > high {
			break
		}
	}

	if wtDistrList[low].sumPrefix >= randomNumber {
		return wtDistrList[low].name
	} else {
		return ""
	}
}

func getCanaryBackend(fnMetadatamap map[string]*metav1.ObjectMeta, fnWtDistributionList []FunctionWeightDistribution) *metav1.ObjectMeta{
	randomNumber := rand.Intn(fnWtDistributionList[len(fnWtDistributionList)].sumPrefix + 1)
	fnName := findCeil(randomNumber,fnWtDistributionList)

	log.Printf("chosen function : %s", fnName)

	return fnMetadatamap[fnName]
}