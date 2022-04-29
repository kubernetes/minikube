include $(sort $(wildcard $(BR2_EXTERNAL_MINIKUBE_PATH)/linux/*.mk))
include $(sort $(wildcard $(BR2_EXTERNAL_MINIKUBE_PATH)/package/*/*.mk))
include $(sort $(wildcard $(BR2_EXTERNAL_MINIKUBE_PATH)/arch/x86_64/package/*/*.mk))
include $(sort $(wildcard $(BR2_EXTERNAL_MINIKUBE_PATH)/arch/aarch64/package/*/*.mk))
