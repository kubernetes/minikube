---
title: "Error Codes"
description: >
  minikube error codes and advice
---



## Generic Errors
1: ExFailure
2: ExInterrupted

## Error codes specific to the minikube program
10: ExProgramError
14: ExProgramUsage
11: ExProgramConflict
15: ExProgramNotFound
16: ExProgramUnsupported
18: ExProgramConfig

## Error codes specific to resource limits (exit code layout follows no rules)
20: ExResourceError
23: ExInsufficientMemory
26: ExInsufficientStorage
27: ExInsufficientPermission
29: ExInsufficientCores

## Error codes specific to the host
30: ExHostError
31: ExHostConflict
32: ExHostTimeout
34: ExHostUsage
35: ExHostNotFound
38: ExHostUnsupported
37: ExHostPermission
38: ExHostConfig

## Error codes specific to remote networking
40: ExInternetError
41: ExInternetConflict
42: ExInternetTimeout
45: ExInternetNotFound
48: ExInternetConfig
49: ExInternetUnavailable

## Error codes specific to the libmachine driver
50: ExDriverError
51: ExDriverConflict
52: ExDriverTimeout
54: ExDriverUsage
55: ExDriverNotFound
56: ExDriverUnsupported
57: ExDriverPermission
58: ExDriverConfig
59: ExDriverUnavailable

## Error codes specific to the driver provider
60: ExProviderError
61: ExProviderConflict
62: ExProviderTimeout
63: ExProviderNotRunning
65: ExProviderNotFound
66: ExProviderUnsupported
67: ExProviderPermission
68: ExProviderConfig
69: ExProviderUnavailable

## Error codes specific to local networking
70: ExLocalNetworkError
71: ExLocalNetworkConflict
72: ExLocalNetworkTimeout
75: ExLocalNetworkNotFound
77: ExLocalNetworkPermission
78: ExLocalNetworkConfig
79: ExLocalNetworkUnavailable

## Error codes specific to the guest host
80: ExGuestError
81: ExGuestConflict
82: ExGuestTimeout
83: ExGuestNotRunning
85: ExGuestNotFound
86: ExGuestUnsupported
87: ExGuestPermission
88: ExGuestConfig
89: ExGuestUnavailable

## Error codes specific to the container runtime
90: ExRuntimeError
93: ExRuntimeNotRunning
95: ExRuntimeNotFound
99: ExRuntimeUnavailable

## Error codes specific to the Kubernetes control plane
100: ExControlPlaneError
101: ExControlPlaneConflict
102: ExControlPlaneTimeout
103: ExControlPlaneNotRunning
105: ExControlPlaneNotFound
106: ExControlPlaneUnsupported
108: ExControlPlaneConfig
109: ExControlPlaneUnavailable

## Error codes specific to a Kubernetes service
110: ExSvcError
111: ExSvcConflict
112: ExSvcTimeout
113: ExSvcNotRunning
115: ExSvcNotFound
116: ExSvcUnsupported
117: ExSvcPermission
118: ExSvcConfig
119: ExSvcUnavailable
