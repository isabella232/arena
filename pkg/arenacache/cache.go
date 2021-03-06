package arenacache

import (
	"fmt"
	"sync"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	pytorchv1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
	tfv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/tensorflow/v1"
	v1 "k8s.io/api/core/v1"
)

var once sync.Once
var arenaCache *ArenaCache

func GetArenaCache() *ArenaCache {
	once.Do(func() {
		locker := new(sync.RWMutex)
		arenaCache = &ArenaCache{
			locker:      locker,
			arenaConfig: config.GetArenaConfiger(),
			pods:        map[string]*v1.Pod{},
			genKey:      func(namespace, name string) string { return fmt.Sprintf("%v/%v", namespace, name) },
		}
	})
	return arenaCache
}

type JobType string
type ArenaCache struct {
	arenaConfig *config.ArenaConfiger
	pods        map[string]*v1.Pod
	tfjobs      map[string]*tfv1.TFJob
	pyjobs      map[string]*pytorchv1.PyTorchJob
	genKey      func(namespace, name string) string
	locker      *sync.RWMutex
}

// AddOrUpdatePod adds or updates pod
func (a *ArenaCache) AddOrUpdatePod(pods ...*v1.Pod) {
	a.locker.Lock()
	defer a.locker.Unlock()
	for _, pod := range pods {
		key := a.genKey(pod.Namespace, pod.Name)
		a.pods[key] = pod.DeepCopy()
	}
}

// DeletePod deletes the pod
func (a *ArenaCache) DeletePod(namespace, name string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	delete(a.pods, a.genKey(namespace, name))
}

func (a *ArenaCache) deletePodsByFilter(filter func(pod *v1.Pod) bool) {
	for _, pod := range a.pods {
		if filter(pod) {
			delete(a.pods, a.genKey(pod.Namespace, pod.Name))
		}
	}
}

// GetPod returns the target pod,if pod not exists,return null
func (a *ArenaCache) GetPod(namespace string, name string) *v1.Pod {
	a.locker.RLock()
	defer a.locker.RUnlock()
	return a.pods[a.genKey(namespace, name)]
}

// FilterPods filter pods that we needs
func (a *ArenaCache) FilterPods(filter func(pod *v1.Pod) bool) []*v1.Pod {
	a.locker.RLock()
	defer a.locker.RUnlock()
	return a.filterPods(filter)
}

// FilterPods filter pods that we needs
func (a *ArenaCache) filterPods(filter func(pod *v1.Pod) bool) []*v1.Pod {
	pods := []*v1.Pod{}
	for _, pod := range a.pods {
		if filter(pod) {
			pods = append(pods, pod)
		}
	}
	return pods
}

// AddOrUpdateJob adds or updates the job,includes: job,tfjob
func (a *ArenaCache) AddOrUpdateJob(job interface{}) {
	a.locker.Lock()
	defer a.locker.Unlock()
	switch v := job.(type) {
	case *tfv1.TFJob:
		a.tfjobs[a.genKey(v.Namespace, v.Name)] = v
	case *pytorchv1.PyTorchJob:
		a.pyjobs[a.genKey(v.Namespace, v.Name)] = v
	}
}

func (a *ArenaCache) DeleteJob(namespace, name string, jobType types.TrainingJobType) {
	a.locker.Lock()
	defer a.locker.Unlock()
	var isMatched func(namespace, name string, pod *v1.Pod) bool
	switch jobType {
	case types.PytorchTrainingJob:
		isMatched = utils.IsPyTorchPod
		delete(a.pyjobs, a.genKey(namespace, name))
	case types.TFTrainingJob:
		isMatched = utils.IsTensorFlowPod
		delete(a.tfjobs, a.genKey(namespace, name))
	default:
		return
	}
	for key, pod := range a.pods {
		if !isMatched(namespace, name, pod) {
			continue
		}
		delete(a.pods, key)
	}
}

// GetAllTFJobs returns all tfjobs and their pods
func (a *ArenaCache) FilterTFJobs(filter func(tfjob *tfv1.TFJob) bool) (map[string]*tfv1.TFJob, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*tfv1.TFJob{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.tfjobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsTensorFlowPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// FilterPytorchJobs returns all tfjobs and their pods
func (a *ArenaCache) FilterPytorchJobs(filter func(pyjob *pytorchv1.PyTorchJob) bool) (map[string]*pytorchv1.PyTorchJob, map[string][]*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	jobs := map[string]*pytorchv1.PyTorchJob{}
	pods := map[string][]*v1.Pod{}
	for jobKey, job := range a.pyjobs {
		if !filter(job) {
			continue
		}
		jobs[jobKey] = job.DeepCopy()
		pods[jobKey] = []*v1.Pod{}
		for _, pod := range a.pods {
			if !utils.IsPyTorchPod(job.Name, job.Namespace, pod) {
				continue
			}
			pods[jobKey] = append(pods[jobKey], pod.DeepCopy())
		}
	}
	return jobs, pods
}

// GetTFJob returns the tfjob
func (a *ArenaCache) GetTFJob(namespace, name string) (*tfv1.TFJob, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.tfjobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsTensorFlowPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}

// get the pytorch job
func (a *ArenaCache) GetPytorchJob(namespace, name string) (*pytorchv1.PyTorchJob, []*v1.Pod) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	pods := []*v1.Pod{}
	job := a.pyjobs[a.genKey(namespace, name)]
	if job == nil {
		return job, pods
	}
	for _, pod := range a.pods {
		if !utils.IsPyTorchPod(job.Name, job.Namespace, pod) {
			continue
		}
		pods = append(pods, pod.DeepCopy())
	}
	return job.DeepCopy(), pods
}
