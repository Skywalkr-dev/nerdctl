/*
   Copyright The containerd Authors.

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

package containerutil

import (
	"reflect"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/labels"
)

func TestParseExtraHosts(t *testing.T) {
	tests := []struct {
		name           string
		extraHosts     []string
		hostGateway    string
		separator      string
		expected       []string
		expectedErrStr string
	}{
		{
			name:     "NoExtraHosts",
			expected: []string{},
		},
		{
			name:       "ExtraHosts",
			extraHosts: []string{"localhost:127.0.0.1", "localhost:[::1]"},
			separator:  ":",
			expected:   []string{"localhost:127.0.0.1", "localhost:[::1]"},
		},
		{
			name:       "EqualsSeperator",
			extraHosts: []string{"localhost:127.0.0.1", "localhost:[::1]"},
			separator:  "=",
			expected:   []string{"localhost=127.0.0.1", "localhost=[::1]"},
		},
		{
			name:           "InvalidExtraHostFormat",
			extraHosts:     []string{"localhost"},
			expectedErrStr: "bad format for add-host: \"localhost\"",
		},
		{
			name:           "ErrorOnHostGatewayExtraHostWithNoHostGatewayIPSet",
			extraHosts:     []string{"localhost:host-gateway"},
			separator:      ":",
			expectedErrStr: "unable to derive the IP value for host-gateway",
		},
		{
			name:        "HostGatewayIP",
			extraHosts:  []string{"localhost:host-gateway"},
			hostGateway: "10.10.0.1",
			separator:   ":",
			expected:    []string{"localhost:10.10.0.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			extraHosts, err := ParseExtraHosts(test.extraHosts, test.hostGateway, test.separator)
			if err != nil && err.Error() != test.expectedErrStr {
				t.Fatalf("expected '%s', actual '%v'", test.expectedErrStr, err)
			} else if err == nil && test.expectedErrStr != "" {
				t.Fatalf("expected error '%s' but got none", test.expectedErrStr)
			}

			if !reflect.DeepEqual(test.expected, extraHosts) {
				t.Fatalf("expected %v, actual %v", test.expected, extraHosts)
			}
		})
	}
}


func TestGetContainerVolumes_Chunked(t *testing.T) {

	part0 := `[{"Type":"volume","Name":"vol-0","Source":"/var/lib/vol-0","Destination":"/mnt/vol-0"},`
	part1 := `{"Type":"volume","Name":"vol-1","Source":"/var/lib/vol-1","Destination":"/mnt/vol-1"},`
	part2 := `{"Type":"volume","Name":"vol-2","Source":"/var/lib/vol-2","Destination":"/mnt/vol-2"}]`
	
	rawJSON := part0 + part1 + part2

	chunkedLabels := map[string]string{
		"nerdctl/mounts/chunk-0": part0,
		"nerdctl/mounts/chunk-1": part1,
		"nerdctl/mounts/chunk-2": part2,
	}

	legacyLabels := map[string]string{
		labels.Mounts: rawJSON,
	}

	chunkedResult := GetContainerVolumes(chunkedLabels)
	legacyResult := GetContainerVolumes(legacyLabels)

	if len(chunkedResult) == 0 {
		t.Fatal("Expected to extract volumes from chunked labels, but got 0 results")
	}

	if len(chunkedResult) != len(legacyResult) {
		t.Errorf("Mismatched output! Chunked found %d volumes, Legacy found %d volumes.", 
			len(chunkedResult), len(legacyResult))
	}

	if chunkedResult[0].Name != "vol-0" {
		t.Errorf("Expected first volume to be named 'vol-0', got '%s'", chunkedResult[0].Name)
	}
	if chunkedResult[2].Name != "vol-2" {
		t.Errorf("Expected third volume to be named 'vol-2', got '%s'", chunkedResult[2].Name)
	}
}
