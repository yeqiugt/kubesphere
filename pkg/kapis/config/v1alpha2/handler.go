package v1alpha2

import (
	"context"
	"errors"
	"regexp"

	"github.com/emicklei/go-restful"
	"github.com/mitchellh/mapstructure"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	ksapi "kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	kserrors "kubesphere.io/kubesphere/pkg/server/errors"
)

const (
	PlatformUIConfigMap = "platform-information"
)

type handler struct {
	k8sClient kubernetes.Interface
}

type PlatformUIConf struct {
	Title       string `json:"title,omitempty" yaml:"title,omitempty" mapstructure:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	Logo        string `json:"logo,omitempty" yaml:"logo,omitempty" mapstructure:"logo"`
	Favicon     string `json:"favicon,omitempty" yaml:"favicon,omitempty" mapstructure:"favicon"`
	Background  string `json:"background,omitempty" yaml:"background,omitempty" mapstructure:"background"`
}

func newPlatformUIHandler(k8sClient kubernetes.Interface) *handler {
	return &handler{k8sClient: k8sClient}
}

func (h handler) createPlatformUI(req *restful.Request, resp *restful.Response) {
	params := PlatformUIConf{}
	err := req.ReadEntity(&params)
	if err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, req, err)
		return
	}

	if !params.Valid() {
		ksapi.HandleBadRequest(resp, req, errors.New("invalid params"))
		return
	}

	configMap := defaultPlatformUICM()
	configMapData := make(map[string]string)
	err = mapstructure.Decode(params, &configMapData)
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, req, err)
		return
	}
	configMap.Data = configMapData
	_, err = h.k8sClient.CoreV1().ConfigMaps(constants.KubeSphereNamespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteAsJson(params)
}

func (h handler) updatePlatformUI(req *restful.Request, resp *restful.Response) {
	params := PlatformUIConf{}
	err := req.ReadEntity(&params)
	if err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, req, err)
		return
	}

	if !params.Valid() {
		ksapi.HandleBadRequest(resp, req, errors.New("invalid params"))
		return
	}
	configMap := defaultPlatformUICM()
	configMapData := make(map[string]string)
	err = mapstructure.Decode(params, &configMapData)
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, req, err)
		return
	}

	configMap.Data = configMapData
	_, err = h.k8sClient.CoreV1().ConfigMaps(constants.KubeSphereNamespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		ksapi.HandleInternalError(resp, req, err)
		return
	}
	resp.WriteAsJson(params)
}

func (h handler) getPlatformUI(req *restful.Request, resp *restful.Response) {
	cm, err := h.k8sClient.CoreV1().ConfigMaps(constants.KubeSphereNamespace).Get(context.TODO(), PlatformUIConfigMap, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		ksapi.HandleNotFound(resp, req, err)
		return
	}
	if cm == nil || cm.Data == nil {
		ksapi.HandleNotFound(resp, req, err)
		return
	}

	resp.WriteAsJson(cm.Data)
}

func (h handler) deletePlatformUI(req *restful.Request, resp *restful.Response) {
	err := h.k8sClient.CoreV1().ConfigMaps(constants.KubeSphereNamespace).Delete(context.TODO(), PlatformUIConfigMap, metav1.DeleteOptions{})
	if err != nil {
		klog.Error(err)
		ksapi.HandleBadRequest(resp, req, err)
		return
	}
	resp.WriteEntity(kserrors.None)
}

func (p *PlatformUIConf) Valid() bool {
	regx := "^[a-z0-9][0-9a-z-]{0,61}[a-z0-9]$"
	compile, err := regexp.Compile(regx)
	if err != nil {
		klog.Warning(err)
		return false
	}
	matchTitle := compile.MatchString(p.Title)
	if !matchTitle {
		return false
	}
	if len([]rune(p.Description)) > 256 {
		return false
	}
	return true
}

func defaultPlatformUICM() *v1.ConfigMap {
	typeMeta := metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}
	objectMeta := metav1.ObjectMeta{Name: PlatformUIConfigMap, Namespace: constants.KubeSphereNamespace}
	return &v1.ConfigMap{TypeMeta: typeMeta, ObjectMeta: objectMeta}
}