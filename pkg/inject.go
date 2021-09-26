package pkg

import (
	"context"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	"github.com/argoproj/gitops-engine/pkg/utils/tracing"
	"github.com/cic-sap/helmize/pkg/utils"
	_ "github.com/cic-sap/helmize/pkg/zerolog"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2/klogr"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"sync/atomic"
)

func Inject(releaseName, chartPath string) error {

	abs, err := filepath.Abs(chartPath)
	if err != nil {
		log.Error().Err(err).Msg("get Error")
		return err
	}
	chartInfo, err := loader.LoadDir(abs)

	restConfig := MustGetRestConfig()
	dynamicClient := dynamic.NewForConfigOrDie(restConfig)

	releaseInfo := &release.Release{
		Name:      releaseName,
		Info:      nil,
		Chart:     nil,
		Config:    nil,
		Manifest:  "",
		Hooks:     nil,
		Version:   0,
		Namespace: os.Getenv("HELM_NAMESPACE"),
		Labels:    nil,
	}

	ctl := &kube.KubectlCmd{
		Log:    klogr.New(),
		Tracer: tracing.NopTracer{},
	}
	ctl.SetOnKubectlRun(func(command string) (kube.CleanupFunc, error) {

		log.Debug().Msgf("run ctl:%s", command)
		return func() {

		}, nil
	})

	h := &HelmInject{
		ctl:        ctl,
		restConfig: restConfig,
		res:        dynamicClient,
		chartPath:  abs,
		chart:      chartInfo,
		release:    releaseInfo,
	}
	res, err := h.loadResource()

	err = h.PatchResource(res)

	if err != nil {
		log.Error().Err(err).Msg("to patch resource get errors")
	} else {
		log.Info().Msg("to patch resource is ok")
	}
	log.Info().Interface("patchResult", h.patchResult).Msg("get patch result")
	return err
}

type HelmInject struct {
	chart       *chart.Chart
	release     *release.Release
	chartPath   string
	res         dynamic.Interface
	restConfig  *rest.Config
	ctl         kube.Kubectl
	patchResult struct {
		Pass    int32 `json:"pass"`
		Error   int32 `json:"error"`
		Success int32 `json:"success"`
		Live    int32 `json:"live"`
		Target  int32 `json:"target"`
		Miss    int32 `json:"miss"`
	}
}

func (h *HelmInject) ns(un *unstructured.Unstructured) string {
	if un.GetNamespace() == "" {
		return h.release.Namespace
	}
	return un.GetNamespace()
}

func (h *HelmInject) patchFunc(un *unstructured.Unstructured) error {

	var err error
	ctx := context.Background()
	var data = map[string]map[string]string{
		"labels": {"app.kubernetes.io/managed-by": "Helm"},
		"annotations": {
			"meta.helm.sh/release-name":      h.release.Name,
			"meta.helm.sh/release-namespace": h.release.Namespace,
		},
	}

	var needPatch = func() bool {
		labels := un.GetLabels()
		if labels == nil {
			return true
		}
		annotations := un.GetAnnotations()
		if annotations == nil {
			return true
		}
		for k, v := range data["labels"] {
			if labels[k] != v {
				return true
			}
		}
		for k, v := range data["annotations"] {
			if annotations[k] != v {
				return true
			}
		}
		return false

	}
	if !needPatch() {
		atomic.AddInt32(&h.patchResult.Pass, 1)
		return nil
	}

	patchBytes, err := json.Marshal(map[string]interface{}{
		"metadata": data,
	})
	if err != nil {
		log.Warn().Err(err).Msg("json get error")
		return err
	}
	log.Debug().Msgf("get patch bytes:%s", string(patchBytes))

	_, err = h.ctl.PatchResource(ctx, h.restConfig, un.GroupVersionKind(), un.GetName(), h.ns(un), types.MergePatchType, patchBytes)

	if err != nil {
		log.Warn().Err(err).Msg("to patch obj get Error")
		atomic.AddInt32(&h.patchResult.Error, 1)
	} else {
		atomic.AddInt32(&h.patchResult.Success, 1)
	}

	return err
}

func (h *HelmInject) loadResource() ([]*unstructured.Unstructured, error) {

	out, _, err := utils.RunShellOutput("helm", []string{"template",
		"-n", h.release.Namespace,
		h.release.Name, h.chartPath})
	if err != nil {
		log.Error().Err(err).Msg("get Error")
		return nil, err
	}
	items, err := kube.SplitYAML([]byte(out))
	return items, err
}

func (h *HelmInject) PatchResource(items []*unstructured.Unstructured) error {

	var err error
	for i, item := range items {
		log.Debug().Msgf("get item:%d, %v", i, kube.GetResourceKey(item))
	}
	h.patchResult.Target = int32(len(items))
	log.Debug().Msgf("get Target:%d", h.patchResult.Target)
	err = kube.RunAllAsync(len(items), func(i int) error {
		targetObj := items[i]
		key := kube.GetResourceKey(targetObj)
		if key.Namespace == "" {
			key.Namespace = h.release.Namespace
		}
		var managedObj *unstructured.Unstructured

		managedObj, err = h.ctl.GetResource(context.TODO(), h.restConfig, targetObj.GroupVersionKind(),
			key.Name, key.Namespace)
		err = client.IgnoreNotFound(err)
		if err != nil {
			log.Warn().Err(err).Msg("get Error")
		}
		if managedObj != nil {
			atomic.AddInt32(&h.patchResult.Live, 1)
			_ = h.patchFunc(managedObj)
		} else {
			atomic.AddInt32(&h.patchResult.Miss, 1)
		}
		return nil
	})

	return err
}

func MustGetRestConfig() *rest.Config {
	conf, err := GetRestConfig()
	if err != nil {
		panic(err)
	}
	return conf
}

func GetRestConfig() (*rest.Config, error) {

	//in cluster
	if os.Getenv("KUBERNETES_PORT") != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		return config, err
	}
	var err error
	// out cluster
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = "~/.kube/config"
	}
	kubeconfig, err = homedir.Expand(kubeconfig)
	if err != nil {
		panic("Error : get env KUBECONFIG")
	}
	os.Setenv("KUBECONFIG", kubeconfig)

	if strings.HasPrefix(kubeconfig, "~") {
		kubeconfig, err = homedir.Expand(kubeconfig)
		if err != nil {
			panic("Error : get env KUBECONFIG")
		}
	}
	if _, err := os.Stat(kubeconfig); err != nil && os.IsNotExist(err) {
		panic(err)
	}
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config, nil

}
