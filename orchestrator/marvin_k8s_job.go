package orchestrator

import (
	"os"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KindJob      = "Job"
	APIVersionV1 = "batch/v1"
)

var (
	// TODO: Make these configurable
	manualSelector           = true
	backoffLimit             = int32(0)
	activeDeadlineSeconds    = int64(300)
	shareProcessNamespace    = true
	uid                      = int64(1000)
	gid                      = int64(3000)
	fsGroup                  = int64(2000)
	runAsNonRoot             = true
	allowPrivilegeEscalation = false
)

type MarvinK8sJob struct {
	JobCreator
}

func (j *MarvinK8sJob) Job() (*batchv1.Job, error) {
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersionV1,
			Kind:       KindJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.Name(),
			Namespace: j.Namespace(),
			Labels:    j.JobLabels(),
		},
		Spec: batchv1.JobSpec{
			ManualSelector:        &manualSelector,
			ActiveDeadlineSeconds: &activeDeadlineSeconds,
			BackoffLimit:          &backoffLimit,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"application": j.Name(),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: j.PodLabels(),
				},
				Spec: corev1.PodSpec{
					ShareProcessNamespace: &shareProcessNamespace,
					Volumes:               j.volumes(),
					SecurityContext:       j.podSecContext(),
					NodeSelector:          j.NodeSelector(),
					ImagePullSecrets:      j.imagePullSecrets(),
					RestartPolicy:         corev1.RestartPolicyNever,
					Containers:            []corev1.Container{*j.container()},
				},
			},
		},
	}
	initContainer := j.initContainer()
	if initContainer != nil {
		job.Spec.Template.Spec.InitContainers = []corev1.Container{*initContainer}
	}
	return job, nil
}

// Container returns a corev1.Container object from the IDriverJob.
func (j *MarvinK8sJob) container() *corev1.Container {
	c := j.Container()
	return &corev1.Container{
		Name:            c.Name,
		Image:           c.Image,
		ImagePullPolicy: corev1.PullPolicy("Always"),
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse(c.Limit.CPU),
				"memory": resource.MustParse(c.Limit.Memory),
			},
		},
		Command:         c.Cmd,
		Args:            c.Args,
		Env:             j.env(c),
		VolumeMounts:    j.mounts(c),
		SecurityContext: j.conSecContext(),
	}
}

// GetInitContainer returns a corev1.Container object from the IDriverJob.
func (j *MarvinK8sJob) initContainer() *corev1.Container {
	c := j.InitContainer()
	if c == nil {
		return nil
	}
	return &corev1.Container{
		Name:            c.Name,
		Image:           c.Image,
		ImagePullPolicy: corev1.PullPolicy("Always"),
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse(c.Limit.CPU),
				"memory": resource.MustParse(c.Limit.Memory),
			},
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse(c.Requests.CPU),
				"memory": resource.MustParse(c.Requests.Memory),
			},
		},
		Command:         c.Cmd,
		Args:            c.Args,
		Env:             j.env(c),
		VolumeMounts:    j.mounts(c),
		SecurityContext: j.conSecContext(),
	}
}

func (j *MarvinK8sJob) volumes() []corev1.Volume {
	var volumes []corev1.Volume

	for _, v := range j.Volumes() {
		volumes = append(volumes, corev1.Volume{
			Name: v,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}

	// Append the volume for mounting the service account for uploading/downloading
	// artifacts from remote storage.
	volumes = append(volumes, corev1.Volume{
		Name: "credentialsdir",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: os.Getenv(EnvNameArtifactsSecretName),
			},
		},
	})
	return volumes
}

func (j *MarvinK8sJob) imagePullSecrets() []corev1.LocalObjectReference {
	var imagePullSecrets []corev1.LocalObjectReference

	for _, v := range j.ImagePullSecrets() {
		imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{
			Name: v,
		})
	}
	return imagePullSecrets
}

func (*MarvinK8sJob) podSecContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    &uid,
		RunAsGroup:   &gid,
		FSGroup:      &fsGroup,
		RunAsNonRoot: &runAsNonRoot,
	}
}

func (*MarvinK8sJob) conSecContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"all",
			},
		},
	}
}

// env is a helper function to convert the IDriverJob's Env to corev1.EnvVars.
func (*MarvinK8sJob) env(c *Container) []corev1.EnvVar {
	var env []corev1.EnvVar

	for k, v := range c.Env {
		env = append(env, corev1.EnvVar{Name: k, Value: v})
	}
	return env
}

// mounts is a helper function to convert the IDriverJob's VolumeMounts to
// corev1.VolumeMounts.
func (*MarvinK8sJob) mounts(c *Container) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	for k, v := range c.VolumeMounts {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      k,
			MountPath: v,
		})
	}
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "credentialsdir",
		MountPath: "/credentials",
	})
	return mounts
}
