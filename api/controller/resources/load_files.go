package resources

import (
	"fmt"
	"net"
	"path"

	"github.com/kubermatic/kubermatic/api"
	"github.com/kubermatic/kubermatic/api/extensions/etcd"
	"github.com/kubermatic/kubermatic/api/provider"
	k8stemplate "github.com/kubermatic/kubermatic/api/template/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	extensionsv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/apis/rbac/v1beta1"
)

// LoadDeploymentFile loads a k8s yaml deployment from disk and returns a Deployment struct
func LoadDeploymentFile(c *api.Cluster, v *api.MasterVersion, masterResourcesPath, dc, yamlFile string) (*extensionsv1beta1.Deployment, error) {
	p, err := provider.ClusterCloudProviderName(c.Spec.Cloud)
	if err != nil {
		return nil, fmt.Errorf("could not identify cloud provider: %v", err)
	}
	data := struct {
		DC               string
		AdvertiseAddress string
		Cluster          *api.Cluster
		Version          *api.MasterVersion
		CloudProvider    string
	}{
		DC:            dc,
		Cluster:       c,
		Version:       v,
		CloudProvider: p,
	}

	addrs, err := net.LookupHost(c.Address.ExternalName)
	if err != nil {
		return nil, err
	}
	data.AdvertiseAddress = addrs[0]

	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, yamlFile))
	if err != nil {
		return nil, err
	}

	var dep extensionsv1beta1.Deployment
	err = t.Execute(data, &dep)
	return &dep, err
}

// LoadServiceFile returns the service for the given cluster and app
func LoadServiceFile(c *api.Cluster, app, masterResourcesPath string) (*v1.Service, error) {
	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, app+"-service.yaml"))
	if err != nil {
		return nil, err
	}

	var service v1.Service

	data := struct {
		Cluster *api.Cluster
	}{
		Cluster: c,
	}

	err = t.Execute(data, &service)

	return &service, err
}

// LoadSecretFile returns the secret for the given cluster and app
func LoadSecretFile(c *api.Cluster, app, masterResourcesPath string) (*v1.Secret, error) {
	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, app+"-secret.yaml"))
	if err != nil {
		return nil, err
	}

	var secret v1.Secret
	data := struct {
		Cluster *api.Cluster
	}{
		Cluster: c,
	}

	err = t.Execute(data, &secret)

	return &secret, err
}

// LoadIngressFile returns the ingress for the given cluster and app
func LoadIngressFile(c *api.Cluster, app, masterResourcesPath string) (*extensionsv1beta1.Ingress, error) {
	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, app+"-ingress.yaml"))
	if err != nil {
		return nil, err
	}
	var ingress extensionsv1beta1.Ingress
	data := struct {
		Cluster *api.Cluster
	}{
		Cluster: c,
	}
	err = t.Execute(data, &ingress)

	if err != nil {
		return nil, err
	}

	return &ingress, err
}

// LoadPVCFile returns the PVC for the given cluster & app
func LoadPVCFile(c *api.Cluster, app, masterResourcesPath string) (*v1.PersistentVolumeClaim, error) {
	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, app+"-pvc.yaml"))
	if err != nil {
		return nil, err
	}

	var pvc v1.PersistentVolumeClaim
	data := struct {
		ClusterName string
	}{
		ClusterName: c.Metadata.Name,
	}
	err = t.Execute(data, &pvc)
	return &pvc, err
}

// LoadAwsCloudConfigConfigMap returns the aws cloud config configMap for the cluster
func LoadAwsCloudConfigConfigMap(c *api.Cluster, dc *provider.DatacenterMeta) (*v1.ConfigMap, error) {
	cm := v1.ConfigMap{}
	cm.Name = "cloud-config"
	cm.APIVersion = "v1"
	cm.Data = map[string]string{
		"config": fmt.Sprintf(`
[global]
zone=%s
VPC=%s
kubernetesclustertag=%s
disablesecuritygroupingress=false
SubnetID=%s
RouteTableID=%s
disablestrictzonecheck=true`, c.Spec.Cloud.AWS.AvailabilityZone, c.Spec.Cloud.AWS.VPCID, c.Metadata.Name, c.Spec.Cloud.AWS.SubnetID, c.Spec.Cloud.AWS.RouteTableID),
	}
	return &cm, nil
}

// LoadOpenstackCloudConfigConfigMap returns the aws cloud config configMap for the cluster
func LoadOpenstackCloudConfigConfigMap(c *api.Cluster, dc *provider.DatacenterMeta) (*v1.ConfigMap, error) {
	//See https://github.com/kubernetes/kubernetes/issues/33128
	config := fmt.Sprintf(`
[Global]
auth-url = "%s"
username = "%s"
password = "%s"
domain-name="%s"
tenant-name = "%s"

[BlockStorage]
trust-device-path = false
`,
		dc.Spec.Openstack.AuthURL,
		c.Spec.Cloud.Openstack.Username,
		c.Spec.Cloud.Openstack.Password,
		c.Spec.Cloud.Openstack.Domain,
		c.Spec.Cloud.Openstack.Tenant,
	)

	cm := v1.ConfigMap{}
	cm.Name = "cloud-config"
	cm.APIVersion = "v1"
	cm.Data = map[string]string{
		"config": config,
	}
	return &cm, nil
}

// LoadEtcdClusterFile loads a etcd-operator tpr from disk and returns a Cluster tpr struct
func LoadEtcdClusterFile(v *api.MasterVersion, masterResourcesPath, yamlFile string) (*etcd.Cluster, error) {

	data := struct {
		Version *api.MasterVersion
	}{
		Version: v,
	}

	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, yamlFile))
	if err != nil {
		return nil, err
	}

	var c etcd.Cluster
	err = t.Execute(data, &c)
	return &c, err
}

// LoadServiceAccountFile loads a service account from disk and returns it
func LoadServiceAccountFile(app, masterResourcesPath string) (*v1.ServiceAccount, error) {
	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, app+"-serviceaccount.yaml"))
	if err != nil {
		return nil, err
	}

	var sa v1.ServiceAccount
	err = t.Execute(nil, &sa)
	return &sa, err
}

// LoadClusterRoleBindingFile loads a role binding from disk, sets the namespace and returns it
func LoadClusterRoleBindingFile(ns, app, masterResourcesPath string) (*v1beta1.ClusterRoleBinding, error) {
	t, err := k8stemplate.ParseFile(path.Join(masterResourcesPath, app+"-rolebinding.yaml"))
	if err != nil {
		return nil, err
	}

	data := struct {
		Namespace string
	}{
		Namespace: ns,
	}

	var r v1beta1.ClusterRoleBinding
	err = t.Execute(data, &r)
	return &r, err
}
