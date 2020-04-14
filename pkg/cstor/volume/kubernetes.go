// Copyright © 2018-2019 The OpenEBS Authors
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

package v1alpha1

import (
	"encoding/json"

	apisv1 "github.com/openebs/api/pkg/apis/cstor/v1"
	clientset "github.com/openebs/api/pkg/client/clientset/versioned"
	client "github.com/openebs/cstor-csi/pkg/kubernetes/client"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getClientsetFn is a typed function that
// abstracts fetching of internal clientset
type getClientsetFn func() (clientset *clientset.Clientset, err error)

// getClientsetFromPathFn is a typed function that
// abstracts fetching of clientset from kubeConfigPath
type getClientsetForPathFn func(
	kubeConfigPath string,
) (*clientset.Clientset, error)

// getFn is a typed function that abstracts get of cstorvolume instances
type getFn func(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apisv1.CStorVolume, error)

// listFn is a typed function that abstracts
// listing of cstor volume instances
type listFn func(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apisv1.CStorVolumeList, error)

// delFn is a typed function that abstracts delete of cstorvolume instances
type delFn func(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error

// createFn is a typed function that abstracts create of cstorvolume instances
type createFn func(
	cli *clientset.Clientset,
	vol *apisv1.CStorVolume,
	namespace string,
) (*apisv1.CStorVolume, error)

// Kubeclient enables kubernetes API operations
// on cstor volume replica instance
type Kubeclient struct {
	// clientset refers to cstor volume replica's
	// clientset that will be responsible to
	// make kubernetes API calls
	clientset      *clientset.Clientset
	kubeConfigPath string
	// namespace holds the namespace on which
	// kubeclient has to operate
	namespace string

	// functions useful during mocking
	getClientset        getClientsetFn
	getClientsetForPath getClientsetForPathFn
	get                 getFn
	list                listFn
	del                 delFn
	create              createFn
}

// KubeclientBuildOption defines the abstraction
// to build a kubeclient instance
type KubeclientBuildOption func(*Kubeclient)

// defaultGetClientset is the default implementation to
// get kubernetes clientset instance
func defaultGetClientset() (clients *clientset.Clientset, err error) {
	config, err := client.GetConfig(client.New())
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(config)
}

// defaultGetClientsetForPath is the default implementation to
// get kubernetes clientset instance based on the given
// kubeconfig path
func defaultGetClientsetForPath(
	kubeConfigPath string,
) (clients *clientset.Clientset, err error) {
	config, err := client.GetConfig(
		client.New(client.WithKubeConfigPath(kubeConfigPath)),
	)
	if err != nil {
		return nil, err
	}
	return clientset.NewForConfig(config)
}

// defaultGet is the default implementation to get a
// cstorvolume instance in kubernetes cluster
func defaultGet(
	cli *clientset.Clientset,
	name, namespace string,
	opts metav1.GetOptions,
) (*apisv1.CStorVolume, error) {
	return cli.CstorV1().
		CStorVolumes(namespace).
		Get(name, opts)
}

// defaultList is the default implementation to list
// cstorvolume instances in kubernetes cluster
func defaultList(
	cli *clientset.Clientset,
	namespace string,
	opts metav1.ListOptions,
) (*apisv1.CStorVolumeList, error) {
	return cli.CstorV1().
		CStorVolumes(namespace).
		List(opts)
}

// defaultDel is the default implementation to delete a
// cstorvolume instance in kubernetes cluster
func defaultDel(
	cli *clientset.Clientset,
	name, namespace string,
	opts *metav1.DeleteOptions,
) error {
	// The object exists in the key-value store until the garbage collector
	// deletes all the dependents whose ownerReference.blockOwnerDeletion=true
	// from the key-value store.  API sever will put the "foregroundDeletion"
	// finalizer on the object, and sets its deletionTimestamp.  This policy is
	// cascading, i.e., the dependents will be deleted with Foreground.
	deletePropagation := metav1.DeletePropagationForeground
	opts.PropagationPolicy = &deletePropagation
	err := cli.CstorV1().
		CStorVolumes(namespace).
		Delete(name, opts)
	return err
}

// defaultCreate is the default implementation to create a
// cstorvolume instance in kubernetes cluster
func defaultCreate(
	cli *clientset.Clientset,
	vol *apisv1.CStorVolume,
	namespace string,
) (*apisv1.CStorVolume, error) {
	return cli.CstorV1().
		CStorVolumes(namespace).
		Create(vol)
}

// withDefaults sets the default options
// of kubeclient instance
func (k *Kubeclient) withDefaults() {
	if k.getClientset == nil {
		k.getClientset = defaultGetClientset
	}

	if k.getClientsetForPath == nil {
		k.getClientsetForPath = defaultGetClientsetForPath
	}

	if k.get == nil {
		k.get = defaultGet
	}

	if k.list == nil {
		k.list = defaultList
	}

	if k.del == nil {
		k.del = defaultDel
	}

	if k.create == nil {
		k.create = defaultCreate
	}
}

// WithClientSet sets the kubernetes client against
// the kubeclient instance
func WithClientSet(c *clientset.Clientset) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.clientset = c
	}
}

// WithNamespace sets the kubernetes client against
// the provided namespace
func WithNamespace(namespace string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.namespace = namespace
	}
}

// WithKubeConfigPath sets the kubernetes client against
// the provided path
func WithKubeConfigPath(path string) KubeclientBuildOption {
	return func(k *Kubeclient) {
		k.kubeConfigPath = path
	}
}

// NewKubeclient returns a new instance of kubeclient meant for
// cstor volume replica operations
func NewKubeclient(opts ...KubeclientBuildOption) *Kubeclient {
	k := &Kubeclient{}
	for _, o := range opts {
		o(k)
	}
	k.withDefaults()
	return k
}

// WithNamespace sets the provided namespace
// against this Kubeclient instance
func (k *Kubeclient) WithNamespace(namespace string) *Kubeclient {
	k.namespace = namespace
	return k
}

func (k *Kubeclient) getClientsetForPathOrDirect() (
	*clientset.Clientset, error) {
	if k.kubeConfigPath != "" {
		return k.getClientsetForPath(k.kubeConfigPath)
	}
	return k.getClientset()
}

// getClientOrCached returns either a new instance
// of kubernetes client or its cached copy
func (k *Kubeclient) getClientOrCached() (*clientset.Clientset, error) {
	if k.clientset != nil {
		return k.clientset, nil
	}
	c, err := k.getClientsetForPathOrDirect()
	if err != nil {
		return nil, err
	}
	k.clientset = c
	return k.clientset, nil
}

// Get returns deployment object for given name
func (k *Kubeclient) Get(
	name string,
	opts metav1.GetOptions,
) (*apisv1.CStorVolume, error) {
	if len(name) == 0 {
		return nil,
			errors.New("failed to get cstorvolume: name can't be empty")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.get(cli, name, k.namespace, opts)
}

// GetRaw returns a raw CStor volume instance
func (k *Kubeclient) GetRaw(
	name string,
	opts metav1.GetOptions,
) ([]byte, error) {
	csv, err := k.Get(name, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(csv)
}

// List returns a list of cstor volume replica
// instances present in kubernetes cluster
func (k *Kubeclient) List(
	opts metav1.ListOptions,
) (*apisv1.CStorVolumeList, error) {
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.list(cli, k.namespace, opts)
}

// Delete deletes the cstorvolume resource
func (k *Kubeclient) Delete(name string) error {
	cli, err := k.getClientOrCached()
	if err != nil {
		return err
	}
	return k.del(cli, name, k.namespace, &metav1.DeleteOptions{})
}

// Create creates the cstorvolume resource
func (k *Kubeclient) Create(vol *apisv1.CStorVolume) (*apisv1.CStorVolume, error) {
	if vol == nil {
		return nil,
			errors.New("failed to create cstorvolume: nil cstorvolume object")
	}
	cli, err := k.getClientOrCached()
	if err != nil {
		return nil, err
	}
	return k.create(cli, vol, k.namespace)
}
