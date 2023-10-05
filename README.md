# RancherSelector
***
RancherSelector is a short program, meant to be run in Rancher downstream clusters.
RancherSelector exposes and API to be consumed by RancherProjector, which can be found at:
https://github.com/wrkode/rancher-projector

## information

RancherProjector will watch the Rancher Management Cluster Projects.
On event such as creation/update/deletion of Projects and Projects Annotations, RancherProjector will hit RancherSelector API on the downstream cluster with POST/Delete requests.
The reasoning behind this is that Rancher Projects do not exist as CRD/Objects in the Rancher Downstream clusters. Therefor (at the time of this writing) there's no way to use the user-defined annotations defined at Project creation.
The counterpart of RancherProjector, RancherSelector, will create a ConfigMap named ```rancher-data``` in the downstream cluster, in Namespace ```kube-system```. this ConfigMap will contain all projects and annotations of all projects of the downstream cluster.

## Usage
**Important**: RancherSelector must be deployed in all downstream clusters, before deploying RancherProjector in the Rancher Management Cluster. 
- clone this repository and ```cd``` into the root directory.
- Deployment file should not need adjusting at this stage, but feel free to customize ```deployment.yaml```, if you know what you're doing.
- Deploy RancherSelector with ```kubectl apply -f deployment.yaml```.
- In namespace ```kube-system```, ConfigMap ```rancher-data``` will be created once RancherProjector is deployed.