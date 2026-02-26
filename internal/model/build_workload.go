package model

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	RegisterKind("Deployment", buildWorkloadInto)
	RegisterKind("StatefulSet", buildWorkloadInto)
	RegisterKind("DaemonSet", buildWorkloadInto)
}

func buildWorkloadInto(m *Model, o unstructured.Unstructured) {
	w := buildWorkload(o)
	kind := o.GetKind()
	m.Workloads = append(m.Workloads, w)
	if kind != "DaemonSet" {
		accumulate(&m.Summary, w)
	}
}

func buildWorkload(o unstructured.Unstructured) Workload {
	ns := nsOrDefault(o.GetNamespace())
	kind := o.GetKind()

	replicas := int64(1)
	if kind == "DaemonSet" {
		replicas = 0
	} else if v, ok, _ := unstructured.NestedInt64(o.Object, "spec", "replicas"); ok {
		replicas = v
	}

	hostNetwork, _, _ := unstructured.NestedBool(o.Object, "spec", "template", "spec", "hostNetwork")
	hostPID, _, _ := unstructured.NestedBool(o.Object, "spec", "template", "spec", "hostPID")
	hostIPC, _, _ := unstructured.NestedBool(o.Object, "spec", "template", "spec", "hostIPC")

	labelMap, _, _ := unstructured.NestedStringMap(o.Object, "spec", "template", "metadata", "labels")
	podLabels := map[string]string{}
	for k, v := range labelMap {
		podLabels[k] = v
	}
	serviceAccountName, _, _ := unstructured.NestedString(o.Object, "spec", "template", "spec", "serviceAccountName")

	podSecCtx := buildPodSecurityContext(o.Object, "spec", "template", "spec", "securityContext")

	volumesAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "template", "spec", "volumes")
	volumes := buildVolumes(volumesAny)

	initContainersAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "template", "spec", "initContainers")
	initContainers := make([]Container, 0, len(initContainersAny))
	for _, cAny := range initContainersAny {
		cMap, ok := cAny.(map[string]any)
		if !ok {
			continue
		}
		initContainers = append(initContainers, buildContainer(cMap))
	}

	containersAny, _, _ := unstructured.NestedSlice(o.Object, "spec", "template", "spec", "containers")
	containers := make([]Container, 0, len(containersAny))
	for _, cAny := range containersAny {
		cMap, ok := cAny.(map[string]any)
		if !ok {
			continue
		}
		containers = append(containers, buildContainer(cMap))
	}

	return Workload{
		Kind:                kind,
		Name:                o.GetName(),
		Namespace:           ns,
		Replicas:            replicas,
		PodLabels:           podLabels,
		ServiceAccountName:  serviceAccountName,
		HostNetwork:         hostNetwork,
		HostPID:             hostPID,
		HostIPC:             hostIPC,
		PodSecurityContext:   podSecCtx,
		Volumes:             volumes,
		InitContainers:      initContainers,
		Containers:          containers,
	}
}

func buildPodSecurityContext(obj map[string]any, path ...string) *PodSecurityContext {
	sec, _, _ := unstructured.NestedMap(obj, path...)
	if len(sec) == 0 {
		return nil
	}
	var runAsUser, runAsGroup, fsGroup *int64
	if v, ok, _ := unstructured.NestedInt64(sec, "runAsUser"); ok {
		runAsUser = &v
	}
	if v, ok, _ := unstructured.NestedInt64(sec, "runAsGroup"); ok {
		runAsGroup = &v
	}
	if v, ok, _ := unstructured.NestedInt64(sec, "fsGroup"); ok {
		fsGroup = &v
	}
	var runAsNonRoot *bool
	if v, ok, _ := unstructured.NestedBool(sec, "runAsNonRoot"); ok {
		runAsNonRoot = &v
	}
	var seccomp *SeccompProfile
	if t, _, _ := unstructured.NestedString(sec, "seccompProfile", "type"); t != "" {
		seccomp = &SeccompProfile{Type: t}
		seccomp.LocalhostProfile, _, _ = unstructured.NestedString(sec, "seccompProfile", "localhostProfile")
	}
	if runAsUser == nil && runAsGroup == nil && runAsNonRoot == nil && fsGroup == nil && seccomp == nil {
		return nil
	}
	return &PodSecurityContext{
		RunAsUser:       runAsUser,
		RunAsGroup:      runAsGroup,
		RunAsNonRoot:    runAsNonRoot,
		FsGroup:         fsGroup,
		SeccompProfile:  seccomp,
	}
}

func buildVolumes(volumesAny []interface{}) []Volume {
	var out []Volume
	for _, vAny := range volumesAny {
		vMap, ok := vAny.(map[string]any)
		if !ok {
			continue
		}
		name, _, _ := unstructured.NestedString(vMap, "name")
		if name == "" {
			continue
		}
		vol := Volume{Name: name}
		if _, ok := vMap["secret"]; ok {
			vol.Type = "secret"
			vol.SecretName, _, _ = unstructured.NestedString(vMap, "secret", "secretName")
		} else if _, ok := vMap["configMap"]; ok {
			vol.Type = "configMap"
			vol.ConfigMapName, _, _ = unstructured.NestedString(vMap, "configMap", "name")
		} else if _, ok := vMap["persistentVolumeClaim"]; ok {
			vol.Type = "persistentVolumeClaim"
			vol.ClaimName, _, _ = unstructured.NestedString(vMap, "persistentVolumeClaim", "claimName")
		} else if _, ok := vMap["emptyDir"]; ok {
			vol.Type = "emptyDir"
		} else if _, ok := vMap["hostPath"]; ok {
			vol.Type = "hostPath"
		} else if _, ok := vMap["projected"]; ok {
			vol.Type = "projected"
		} else {
			vol.Type = "unknown"
		}
		out = append(out, vol)
	}
	return out
}

func buildContainer(cMap map[string]any) Container {
	name, _, _ := unstructured.NestedString(cMap, "name")
	image, _, _ := unstructured.NestedString(cMap, "image")
	imagePullPolicy, _, _ := unstructured.NestedString(cMap, "imagePullPolicy")
	c := Container{Name: name, Image: image, ImagePullPolicy: imagePullPolicy}

	mountsAny, _, _ := unstructured.NestedSlice(cMap, "volumeMounts")
	for _, mAny := range mountsAny {
		mMap, ok := mAny.(map[string]any)
		if !ok {
			continue
		}
		mountName, _, _ := unstructured.NestedString(mMap, "name")
		mountPath, _, _ := unstructured.NestedString(mMap, "mountPath")
		readOnly, _, _ := unstructured.NestedBool(mMap, "readOnly")
		c.VolumeMounts = append(c.VolumeMounts, VolumeMount{Name: mountName, MountPath: mountPath, ReadOnly: readOnly})
	}

	envRefs, secretRefs := buildEnvRefs(cMap)
	c.ConfigMapRefs = envRefs
	c.SecretRefs = secretRefs

	if _, ok, _ := unstructured.NestedMap(cMap, "readinessProbe"); ok {
		c.HasReadiness = true
	}
	if _, ok, _ := unstructured.NestedMap(cMap, "livenessProbe"); ok {
		c.HasLiveness = true
	}
	if _, ok, _ := unstructured.NestedMap(cMap, "startupProbe"); ok {
		c.HasStartup = true
	}

	sec, _, _ := unstructured.NestedMap(cMap, "securityContext")
	if len(sec) > 0 {
		if v, ok, _ := unstructured.NestedBool(sec, "privileged"); ok && v {
			c.Privileged = true
		}
		if v, ok, _ := unstructured.NestedBool(sec, "readOnlyRootFilesystem"); ok && v {
			c.ReadOnlyRootFilesystem = true
		}
		if v, ok, _ := unstructured.NestedBool(sec, "allowPrivilegeEscalation"); ok {
			c.AllowPrivilegeEscalation = &v
		}
		if v, ok, _ := unstructured.NestedBool(sec, "runAsNonRoot"); ok {
			c.RunAsNonRoot = &v
		}
	}

	c.CPURequest = nestedQuantity(cMap, "resources", "requests", "cpu")
	c.MemRequest = nestedQuantity(cMap, "resources", "requests", "memory")
	c.CPULimit = nestedQuantity(cMap, "resources", "limits", "cpu")
	c.MemLimit = nestedQuantity(cMap, "resources", "limits", "memory")
	return c
}

func buildEnvRefs(cMap map[string]any) (configMapRefs, secretRefs []string) {
	seenCM := map[string]bool{}
	seenSec := map[string]bool{}
	envAny, _, _ := unstructured.NestedSlice(cMap, "env")
	for _, eAny := range envAny {
		eMap, ok := eAny.(map[string]any)
		if !ok {
			continue
		}
		if name, _, _ := unstructured.NestedString(eMap, "valueFrom", "configMapKeyRef", "name"); name != "" && !seenCM[name] {
			seenCM[name] = true
			configMapRefs = append(configMapRefs, name)
		}
		if name, _, _ := unstructured.NestedString(eMap, "valueFrom", "secretKeyRef", "name"); name != "" && !seenSec[name] {
			seenSec[name] = true
			secretRefs = append(secretRefs, name)
		}
	}
	envFromAny, _, _ := unstructured.NestedSlice(cMap, "envFrom")
	for _, eAny := range envFromAny {
		eMap, ok := eAny.(map[string]any)
		if !ok {
			continue
		}
		if name, _, _ := unstructured.NestedString(eMap, "configMapRef", "name"); name != "" && !seenCM[name] {
			seenCM[name] = true
			configMapRefs = append(configMapRefs, name)
		}
		if name, _, _ := unstructured.NestedString(eMap, "secretRef", "name"); name != "" && !seenSec[name] {
			seenSec[name] = true
			secretRefs = append(secretRefs, name)
		}
	}
	return configMapRefs, secretRefs
}

func accumulate(sum *Summary, w Workload) {
	for _, c := range w.Containers {
		if c.CPURequest != nil {
			sum.CPURequests.Add(*c.CPURequest)
		}
		if c.CPULimit != nil {
			sum.CPULimits.Add(*c.CPULimit)
		}
		if c.MemRequest != nil {
			sum.MemRequests.Add(*c.MemRequest)
		}
		if c.MemLimit != nil {
			sum.MemLimits.Add(*c.MemLimit)
		}
	}
}
