// Copyright 2016 CoreOS, Inc.
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

package cluster

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/coreos/mantle/harness"
	"github.com/coreos/mantle/platform"
)

// TestCluster embedds a Cluster to provide platform independant helper
// methods.
type TestCluster struct {
	*harness.H
	platform.Cluster
	NativeFuncs []string
}

// Run runs f as a subtest and reports whether f succeeded.
func (t *TestCluster) Run(name string, f func(c TestCluster)) bool {
	return t.H.Run(name, func(h *harness.H) {
		f(TestCluster{H: h, Cluster: t.Cluster})
	})
}

// RunNative runs a registered NativeFunc on a remote machine
func (t *TestCluster) RunNative(funcName string, m platform.Machine) bool {
	command := fmt.Sprintf("./kolet run %q %q", t.Name(), funcName)
	return t.Run(funcName, func(c TestCluster) {
		client, err := m.SSHClient()
		if err != nil {
			c.Fatalf("kolet SSH client: %v", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			c.Fatalf("kolet SSH session: %v", err)
		}
		defer session.Close()

		b, err := session.CombinedOutput(command)
		b = bytes.TrimSpace(b)
		if len(b) > 0 {
			t.Logf("kolet:\n%s", b)
		}
		if err != nil {
			c.Errorf("kolet: %v", err)
		}
	})
}

// ListNativeFunctions returns a slice of function names that can be executed
// directly on machines in the cluster.
func (t *TestCluster) ListNativeFunctions() []string {
	return t.NativeFuncs
}

// DropFile places file from localPath to ~/ on every machine in cluster
func (t *TestCluster) DropFile(localPath string) error {
	in, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer in.Close()

	for _, m := range t.Machines() {
		if _, err := in.Seek(0, 0); err != nil {
			return err
		}
		if err := platform.InstallFile(in, m, filepath.Base(localPath)); err != nil {
			return err
		}
	}
	return nil
}
