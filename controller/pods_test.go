// Copyright 2017 Cisco Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"
	"time"

	"github.com/noironetworks/aci-containers/metadata"
	tu "github.com/noironetworks/aci-containers/testutil"
)

type annotTest struct {
	ns      string
	egannot string
	sgannot string
	desc    string
}

const egAnnot = "{\"policy-space\": \"testps\", \"name\": \"test|test-eg\"}"
const sgAnnot = "[{\"policy-space\": \"testps\", \"name\": \"test-sg\"}]"

var annotTests = []annotTest{
	{"testns", egAnnot, "", "egonly"},
	{"testns", egAnnot, sgAnnot, "both"},
	{"testns", "", sgAnnot, "sgonly"},
	{"testns", "", "", "neither"},
}

func waitForGroupAnnot(t *testing.T, cont *testAciController,
	egAnnot string, sgAnnot string, desc string) {
	tu.WaitFor(t, desc, 500*time.Millisecond,
		func(last bool) (bool, error) {
			if !tu.WaitCondition(t, last, func() bool {
				return len(cont.podUpdates) >= 1
			}, desc) {
				return false, nil
			}
			annot :=
				cont.podUpdates[len(cont.podUpdates)-1].ObjectMeta.Annotations
			return tu.WaitEqual(t, last, egAnnot,
				annot[metadata.CompEgAnnotation], desc) &&
				tu.WaitEqual(t, last, sgAnnot,
					annot[metadata.CompSgAnnotation], desc), nil
		})
}

func TestPodAnnotation(t *testing.T) {
	cont := testController()
	cont.run()

	for _, test := range annotTests {
		cont.podUpdates = nil
		cont.fakePodSource.Add(pod("testpod", test.ns, test.egannot, test.sgannot))
		waitForGroupAnnot(t, cont, test.egannot, test.sgannot, test.desc)
	}

	cont.stop()
}

func TestNamespaceAnnotation(t *testing.T) {
	cont := testController()
	cont.run()

	cont.fakePodSource.Add(pod("testns", "testpod", "", ""))
	waitForGroupAnnot(t, cont, "", "", "none")

	for _, test := range annotTests {
		cont.podUpdates = nil
		cont.fakeNamespaceSource.Add(namespace(test.ns, test.egannot, test.sgannot))
		waitForGroupAnnot(t, cont, test.egannot, test.sgannot, test.desc)
	}

	cont.fakeNamespaceSource.Delete(namespace("testns", "", ""))
	waitForGroupAnnot(t, cont, "", "", "none")

	cont.stop()
}

func TestDeploymentAnnotation(t *testing.T) {
	cont := testController()
	cont.run()

	cont.fakePodSource.Add(pod("testns", "testpod", "", ""))
	waitForGroupAnnot(t, cont, "", "", "none")

	for _, test := range annotTests {
		cont.podUpdates = nil
		cont.fakeDeploymentSource.Add(deployment(test.ns, "testdep", test.egannot, test.sgannot))
		waitForGroupAnnot(t, cont, test.egannot, test.sgannot, test.desc)
	}

	cont.fakeDeploymentSource.Delete(deployment("testns", "testdep", "", ""))
	waitForGroupAnnot(t, cont, "", "", "none")

	cont.stop()
}

func TestDeploymentAnnotationPre(t *testing.T) {
	cont := testController()
	cont.run()

	cont.fakePodSource.Add(pod("testns", "testpod", "", ""))
	waitForGroupAnnot(t, cont, "", "", "pod1")

	cont.podUpdates = nil
	cont.fakeDeploymentSource.Add(deployment("testns", "testdep", egAnnot, sgAnnot))
	waitForGroupAnnot(t, cont, egAnnot, sgAnnot, "pod1update")

	cont.podUpdates = nil
	cont.fakePodSource.Add(pod("testns", "testpod2", "", ""))
	waitForGroupAnnot(t, cont, egAnnot, sgAnnot, "pod2")

	cont.stop()
}