/*
Copyright 2022.

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

package v1

import (
	"strings"

	"github.com/linkall-labs/vanus-operator/internal/status"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VanusClusterSpec defines the desired state of VanusCluster
type VanusClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is the number of nodes in the RabbitMQ cluster. Each node is deployed as a Replica in a StatefulSet. Only 1, 3, 5 replicas clusters are tested.
	// This value should be an odd number to ensure the resultant cluster can establish exactly one quorum of nodes
	// in the event of a fragmenting network partition.
	// +optional
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`
	// The desired compute resource requirements of Pods in the cluster.
	// +kubebuilder:default:={limits: {cpu: "2000m", memory: "2Gi"}, requests: {cpu: "1000m", memory: "2Gi"}}
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Affinity scheduling rules to be applied on created Pods.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// Tolerations is the list of Toleration resources attached to each Pod in the VanusCluster.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Provides the ability to override the generated manifest of several child resources.
	Override VanusClusterOverrideSpec `json:"override,omitempty"`
	// TerminationGracePeriodSeconds is the timeout that each rabbitmqcluster pod will have to terminate gracefully.
	// It defaults to 604800 seconds ( a week long) to ensure that the container preStop lifecycle hook can finish running.
	// For more information, see: https://github.com/rabbitmq/cluster-operator/blob/main/docs/design/20200520-graceful-pod-termination.md
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=604800
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// DelayStartSeconds is the time the init container (`setup-container`) will sleep before terminating.
	// This effectively delays the time between starting the Pod and starting the `rabbitmq` container.
	// RabbitMQ relies on up-to-date DNS entries early during peer discovery.
	// The purpose of this artificial delay is to ensure that DNS entries are up-to-date when booting RabbitMQ.
	// For more information, see https://github.com/kubernetes/kubernetes/issues/92559
	// If your Kubernetes DNS backend is configured with a low DNS cache value or publishes not ready addresses
	// promptly, you can decrase this value or set it to 0.
	// +kubebuilder:validation:Minimum:=0
	// +kubebuilder:default:=30
	DelayStartSeconds *int32 `json:"delayStartSeconds,omitempty"`
}

// VanusClusterStatus defines the observed state of VanusCluster
type VanusClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Set of Conditions describing the current state of the VanusCluster
	Conditions []status.VanusClusterCondition `json:"conditions"`

	// Identifying information on internal resources
	DefaultUser *VanusClusterDefaultUser `json:"defaultUser,omitempty"`

	// Binding exposes a secret containing the binding information for this
	// VanusCluster. It implements the service binding Provisioned Service
	// duck type. See: https://github.com/servicebinding/spec#provisioned-service
	Binding *corev1.LocalObjectReference `json:"binding,omitempty"`

	// observedGeneration is the most recent successful generation observed for this VanusCluster. It corresponds to the
	// VanusCluster's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VanusCluster is the Schema for the vanusclusters API
type VanusCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VanusClusterSpec   `json:"spec,omitempty"`
	Status VanusClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VanusClusterList contains a list of VanusCluster
type VanusClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VanusCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VanusCluster{}, &VanusClusterList{})
}

// Provides the ability to override the generated manifest of several child resources.
type VanusClusterOverrideSpec struct {
	// Override configuration for the RabbitMQ StatefulSet.
	Controller *StatefulSet `json:"controller,omitempty"`
	// Override configuration for the RabbitMQ StatefulSet.
	Store *StatefulSet `json:"store,omitempty"`
	// Override configuration for the RabbitMQ StatefulSet.
	Gateway *Deployment `json:"gateway,omitempty"`
	// Override configuration for the RabbitMQ StatefulSet.
	Trigger *Deployment `json:"trigger,omitempty"`
	// Override configuration for the RabbitMQ StatefulSet.
	Timer *Deployment `json:"timer,omitempty"`
}

// Override configuration for the RabbitMQ StatefulSet.
// Allows for the manifest of the created StatefulSet to be overwritten with custom configuration.
type StatefulSet struct {
	// +optional
	*EmbeddedLabelsAnnotations `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Spec defines the desired identities of pods in this set.
	// +optional
	Spec *StatefulSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// StatefulSetSpec contains a subset of the fields included in k8s.io/api/apps/v1.StatefulSetSpec.
// Field RevisionHistoryLimit is omitted.
// Every field is made optional.
type StatefulSetSpec struct {
	// replicas corresponds to the desired number of Pods in the StatefulSet.
	// For more info, see https://pkg.go.dev/k8s.io/api/apps/v1#StatefulSetSpec
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	// template is the object that describes the pod that will be created if
	// insufficient replicas are detected. Each pod stamped out by the StatefulSet
	// will fulfill this Template, but have a unique identity from the rest
	// of the StatefulSet.
	// +optional
	Template *PodTemplateSpec `json:"template,omitempty" protobuf:"bytes,3,opt,name=template"`

	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// +optional
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`

	// serviceName is the name of the service that governs this StatefulSet.
	// This service must exist before the StatefulSet, and is responsible for
	// the network identity of the set. Pods get DNS/hostnames that follow the
	// pattern: pod-specific-string.serviceName.default.svc.cluster.local
	// where "pod-specific-string" is managed by the StatefulSet controller.
	// +optional
	ServiceName string `json:"serviceName,omitempty" protobuf:"bytes,5,opt,name=serviceName"`

	// podManagementPolicy controls how pods are created during initial scale up,
	// when replacing pods on nodes, or when scaling down. The default policy is
	// `OrderedReady`, where pods are created in increasing order (pod-0, then
	// pod-1, etc) and the controller will wait until each pod is ready before
	// continuing. When scaling down, the pods are removed in the opposite order.
	// The alternative policy is `Parallel` which will create pods in parallel
	// to match the desired scale without waiting, and on scale down will delete
	// all pods at once.
	// +optional
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty" protobuf:"bytes,6,opt,name=podManagementPolicy,casttype=PodManagementPolicyType"`

	// updateStrategy indicates the StatefulSetUpdateStrategy that will be
	// employed to update Pods in the StatefulSet when a revision is made to
	// Template.
	// +optional
	UpdateStrategy *appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty" protobuf:"bytes,7,opt,name=updateStrategy"`

	// The minimum number of seconds for which a newly created StatefulSet pod should
	// be ready without any of its container crashing, for it to be considered
	// available. Defaults to 0 (pod will be considered available as soon as it
	// is ready).
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,4,opt,name=minReadySeconds"`
}

// Override configuration for the RabbitMQ StatefulSet.
// Allows for the manifest of the created StatefulSet to be overwritten with custom configuration.
type Deployment struct {
	// +optional
	*EmbeddedLabelsAnnotations `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Spec defines the desired identities of pods in this set.
	// +optional
	Spec *DeploymentSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// StatefulSetSpec contains a subset of the fields included in k8s.io/api/apps/v1.StatefulSetSpec.
// Field RevisionHistoryLimit is omitted.
// Every field is made optional.
type DeploymentSpec struct {
	// replicas corresponds to the desired number of Pods in the StatefulSet.
	// For more info, see https://pkg.go.dev/k8s.io/api/apps/v1#StatefulSetSpec
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	// template is the object that describes the pod that will be created if
	// insufficient replicas are detected. Each pod stamped out by the StatefulSet
	// will fulfill this Template, but have a unique identity from the rest
	// of the StatefulSet.
	// +optional
	Template *PodTemplateSpec `json:"template,omitempty" protobuf:"bytes,3,opt,name=template"`

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	// +patchStrategy=retainKeys
	Strategy appsv1.DeploymentStrategy `json:"strategy,omitempty" patchStrategy:"retainKeys" protobuf:"bytes,4,opt,name=strategy"`

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,5,opt,name=minReadySeconds"`

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	// Indicates that the deployment is paused.
	// +optional
	Paused bool `json:"paused,omitempty" protobuf:"varint,7,opt,name=paused"`

	// The maximum time in seconds for a deployment to make progress before it
	// is considered to be failed. The deployment controller will continue to
	// process failed deployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the deployment status. Note that progress will
	// not be estimated during the time a deployment is paused. Defaults to 600s.
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,9,opt,name=progressDeadlineSeconds"`
}

// EmbeddedLabelsAnnotations is an embedded subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta.
// Only labels and annotations are included.
type EmbeddedLabelsAnnotations struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

// EmbeddedObjectMeta is an embedded subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta.
// Only fields which are relevant to embedded resources are included.
type EmbeddedObjectMeta struct {
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Namespace defines the space within each name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	//
	// Must be a DNS_LABEL.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/namespaces
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

// PodTemplateSpec is an embedded version of k8s.io/api/core/v1.PodTemplateSpec.
// It contains a reduced ObjectMeta.
type PodTemplateSpec struct {
	// +optional
	*EmbeddedObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec *corev1.PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// PersistentVolumeClaim is an embedded version of k8s.io/api/core/v1.PersistentVolumeClaim.
// It contains TypeMeta and a reduced ObjectMeta.
// Field status is omitted.
type PersistentVolumeClaim struct {
	// Embedded metadata identifying a Kind and API Version of an object.
	// For more info, see: https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#TypeMeta
	metav1.TypeMeta `json:",inline"`
	// +optional
	EmbeddedObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired characteristics of a volume requested by a pod author.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	Spec corev1.PersistentVolumeClaimSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// Contains references to resources created with the VanusCluster resource.
type VanusClusterDefaultUser struct {
	// Reference to the Kubernetes Secret containing the credentials of the default
	// user.
	SecretReference *VanusClusterSecretReference `json:"secretReference,omitempty"`
	// Reference to the Kubernetes Service serving the cluster.
	ServiceReference *VanusClusterServiceReference `json:"serviceReference,omitempty"`
}

// Reference to the Kubernetes Secret containing the credentials of the default user.
type VanusClusterSecretReference struct {
	// Name of the Secret containing the default user credentials
	Name string `json:"name"`
	// Namespace of the Secret containing the default user credentials
	Namespace string `json:"namespace"`
	// Key-value pairs in the Secret corresponding to `username`, `password`, `host`, and `port`
	Keys map[string]string `json:"keys"`
}

// Reference to the Kubernetes Service serving the cluster.
type VanusClusterServiceReference struct {
	// Name of the Service serving the cluster
	Name string `json:"name"`
	// Namespace of the Service serving the cluster
	Namespace string `json:"namespace"`
}

func (cluster VanusCluster) ChildResourceName(name string) string {
	return strings.TrimSuffix(strings.Join([]string{cluster.Name, name}, "-"), "-")
}

func (clusterStatus *VanusClusterStatus) SetConditions(resources []runtime.Object) {
	var oldAllPodsReadyCondition *status.VanusClusterCondition
	var oldClusterAvailableCondition *status.VanusClusterCondition
	var oldNoWarningsCondition *status.VanusClusterCondition
	var oldReconcileCondition *status.VanusClusterCondition

	for _, condition := range clusterStatus.Conditions {
		switch condition.Type {
		case status.AllReplicasReady:
			oldAllPodsReadyCondition = condition.DeepCopy()
		case status.ClusterAvailable:
			oldClusterAvailableCondition = condition.DeepCopy()
		case status.NoWarnings:
			oldNoWarningsCondition = condition.DeepCopy()
		case status.ReconcileSuccess:
			oldReconcileCondition = condition.DeepCopy()
		}
	}

	allReplicasReadyCond := status.AllReplicasReadyCondition(resources, oldAllPodsReadyCondition)
	clusterAvailableCond := status.ClusterAvailableCondition(resources, oldClusterAvailableCondition)
	noWarningsCond := status.NoWarningsCondition(resources, oldNoWarningsCondition)

	var reconciledCondition status.VanusClusterCondition
	if oldReconcileCondition != nil {
		reconciledCondition = *oldReconcileCondition
	} else {
		reconciledCondition = status.ReconcileSuccessCondition(corev1.ConditionUnknown, "Initialising", "")
	}

	clusterStatus.Conditions = []status.VanusClusterCondition{
		allReplicasReadyCond,
		clusterAvailableCond,
		noWarningsCond,
		reconciledCondition,
	}
}

func (clusterStatus *VanusClusterStatus) SetCondition(condType status.VanusClusterConditionType,
	condStatus corev1.ConditionStatus, reason string, messages ...string) {
	for i := range clusterStatus.Conditions {
		if clusterStatus.Conditions[i].Type == condType {
			clusterStatus.Conditions[i].UpdateState(condStatus)
			clusterStatus.Conditions[i].UpdateReason(reason, messages...)
			break
		}
	}
}
