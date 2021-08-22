/*
Copyright 2021.

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

package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,sideEffects=noneOnDryRun,admissionReviewVersions=v1,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io

// gatekeeperInjector injects sidecars
type gatekeeperInjector struct {
	Client  client.Client
	decoder *admission.Decoder
}

// log is for logging in this package.
var gatekeeperInjectorLog = logf.Log.WithName("gatekeeperInjector")

func NewGatekeeperInjector(c client.Client) admission.Handler {
	return &gatekeeperInjector{Client: c}
}

// gatekeeperInjector adds an annotation to every incoming pods.
func (a *gatekeeperInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := a.decoder.Decode(req, pod)

	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	annotations := pod.GetAnnotations()
	if _, ok := annotations["gatekeeper.gogatekeeper"]; !ok {
		return admission.Allowed("No injection requested")
	}

	configMapVolume := &corev1.ConfigMapVolumeSource{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: annotations["gatekeeper.gogatekeeper"],
		},
	}

	gatekeeperConfigVolume := corev1.Volume{
		Name: "gatekeeper-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: configMapVolume,
		},
	}

	gatekeeperContainer := corev1.Container{
		Image: "quay.io/gogatekeeper/gatekeeper:1.3.4",
		Name:  "gogatekeeper",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "gatekeeper-config",
				MountPath: "/etc/gatekeeperConfig/",
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "gatekeeper",
				ContainerPort: 3000,
			},
		},
		Args: []string{
			"--config",
			"/etc/gatekeeperConfig/gatekeeper.yaml",
			"--client-id",
			annotations["gatekeeper.gogatekeeper/client-id"],
			"--client-secret",
			annotations["gatekeeper.gogatekeeper/client-secret"],
			"--encryption-key",
			annotations["gatekeeper.gogatekeeper/encryption-key"],
			"--redirection-url",
			annotations["gatekeeper.gogatekeeper/redirection-url"],
		},
	}

	pod.Spec.Containers = append(pod.Spec.Containers, gatekeeperContainer)
	pod.Spec.Volumes = append(pod.Spec.Volumes, gatekeeperConfigVolume)

	gatekeeperInjectorLog.Info("incoming pod", "annotations", pod.GetAnnotations())

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// gatekeeperInjector implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *gatekeeperInjector) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
