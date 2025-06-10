package environment

import (
	"io"
	"k8s.io/minikube/pkg/minikube/shell"
)

// EnvConfigurator 是任何环境配置逻辑都必须实现的接口。
type EnvConfigurator interface {
	// Vars 返回需要被设置的环境变量的 map。
	Vars() (map[string]string, error)

	// UnsetVars 返回需要被取消设置的环境变量名称列表。
	UnsetVars() ([]string, error)

	// DisplayScript 生成需要显示的、用于设置环境变量的脚本。
	DisplayScript(sh shell.Config, w io.Writer) error
}