// Copyright 2026 Silence-Operator Maintainers
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
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/silence-operator/silence-operator/internal/tlsutil"
)

// runnableAdder is the subset of ctrl.Manager needed to unit test addCertWatcher with a stub.
type runnableAdder interface {
	Add(manager.Runnable) error
}

// setupCertWatcher wraps tlsutil.Watcher for component ("webhook" or "metrics"), logging when it finds a certificate.
func setupCertWatcher(component, certPath, certName, keyName string) (*certwatcher.CertWatcher, error) {
	watcher, err := tlsutil.Watcher(certPath, certName, keyName)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize %s certificate watcher: %w", component, err)
	}
	if watcher != nil {
		setupLog.Info(fmt.Sprintf("Initializing %s certificate watcher using provided certificates", component),
			component+"-cert-path", certPath, component+"-cert-name", certName, component+"-cert-key", keyName)
	}
	return watcher, nil
}

// addCertWatcher registers a non-nil watcher on adder under component's name, logging as it does so.
func addCertWatcher(adder runnableAdder, component string, watcher *certwatcher.CertWatcher) error {
	if watcher == nil {
		return nil
	}
	setupLog.Info(fmt.Sprintf("Adding %s certificate watcher to manager", component))
	if err := adder.Add(watcher); err != nil {
		return fmt.Errorf("unable to add %s certificate watcher to manager: %w", component, err)
	}
	return nil
}
