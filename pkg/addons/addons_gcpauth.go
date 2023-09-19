/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package addons

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	credentialsPath = "/var/lib/minikube/google_application_credentials.json"
	projectPath     = "/var/lib/minikube/google_cloud_project"
	secretName      = "gcp-auth"
	namespaceName   = "gcp-auth"

	// readPermission correlates to read-only file system permissions
	readPermission = "0444"
)

// enableOrDisableGCPAuth enables or disables the gcp-auth addon depending on the val parameter
func enableOrDisableGCPAuth(cfg *config.ClusterConfig, name, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddonGCPAuth(cfg)
	}
	return disableAddonGCPAuth(cfg)
}

func enableAddonGCPAuth(cfg *config.ClusterConfig) error {
	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Grab credentials from where GCP would normally look
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		if detect.IsCloudShell() {
			if c := os.Getenv("CLOUDSDK_CONFIG"); c != "" {
				f, err := os.ReadFile(path.Join(c, "application_default_credentials.json"))
				if err == nil {
					creds, _ = google.CredentialsFromJSON(ctx, f)
				}
			}
		} else {
			exit.Message(reason.InternalCredsNotFound, "Could not find any GCP credentials. Either run `gcloud auth application-default login` or set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the path of your credentials file.")
		}
	}

	// Patch service accounts for all namespaces to include the image pull secret.
	// The image registry pull secret is added to the namespaces in the webhook.
	if err := patchServiceAccounts(cfg); err != nil {
		return errors.Wrap(err, "patching service accounts")
	}

	// If the env var is explicitly set, even in GCE, then defer to the user and continue
	if !Force && detect.IsOnGCE() && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		out.WarningT("It seems that you are running in GCE, which means authentication should work without the GCP Auth addon. If you would still like to authenticate using a credentials file, use the --force flag.")
		return nil
	}

	if creds.JSON == nil {
		out.WarningT("You have authenticated with a service account that does not have an associated JSON file. The GCP Auth addon requires credentials with a JSON file in order to continue.")
		return nil
	}

	// Actually copy the creds over
	f := assets.NewMemoryAssetTarget(creds.JSON, credentialsPath, readPermission)

	if err := r.Copy(f); err != nil {
		return err
	}

	// First check if the project env var is explicitly set
	projectEnv := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectEnv != "" {
		f := assets.NewMemoryAssetTarget([]byte(projectEnv), projectPath, readPermission)
		return r.Copy(f)
	}

	// We're currently assuming gcloud is installed and in the user's path
	proj, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err == nil && len(proj) > 0 {
		f := assets.NewMemoryAssetTarget(bytes.TrimSpace(proj), projectPath, readPermission)
		return r.Copy(f)
	}

	out.WarningT("Could not determine a Google Cloud project, which might be ok.")
	out.Styled(style.Tip, `To set your Google Cloud project,  run:

		gcloud config set project <project name>

or set the GOOGLE_CLOUD_PROJECT environment variable.`)

	// Copy an empty file in to avoid errors about missing files
	emptyFile := assets.NewMemoryAssetTarget([]byte{}, projectPath, readPermission)
	return r.Copy(emptyFile)

}

func patchServiceAccounts(cc *config.ClusterConfig) error {
	client, err := service.K8s.GetCoreClient(cc.Name)
	if err != nil {
		return err
	}

	namespaces, err := client.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, n := range namespaces.Items {
		// Now patch the secret into all the service accounts we can find
		serviceaccounts := client.ServiceAccounts(n.Name)
		salist, err := serviceaccounts.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		// Let's make sure we at least find the default service account
		for len(salist.Items) == 0 {
			salist, err = serviceaccounts.List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
		}

		ips := corev1.LocalObjectReference{Name: secretName}
		for _, sa := range salist.Items {
			add := true
			for _, ps := range sa.ImagePullSecrets {
				if ps.Name == secretName {
					add = false
					break
				}
			}
			if add {
				sa.ImagePullSecrets = append(sa.ImagePullSecrets, ips)
				_, err := serviceaccounts.Update(context.TODO(), &sa, metav1.UpdateOptions{})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func refreshExistingPods(cc *config.ClusterConfig) error {
	klog.Info("refreshing existing pods")
	client, err := service.K8s.GetCoreClient(cc.Name)
	if err != nil {
		return fmt.Errorf("failed to get k8s client: %v", err)
	}

	namespaces, err := client.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get namespaces: %v", err)
	}
	for _, n := range namespaces.Items {
		// Ignore kube-system and gcp-auth namespaces
		if skipNamespace(n.Name) {
			continue
		}

		pods := client.Pods(n.Name)
		podList, err := pods.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}

		for _, p := range podList.Items {
			// Skip pods we're explicitly told to skip
			if _, ok := p.Labels["gcp-auth-skip-secret"]; ok {
				continue
			}

			klog.Infof("refreshing pod %q", p.Name)

			// Recreating the pod should pickup the necessary changes
			err := pods.Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete pod %q: %v", p.Name, err)
			}

			p.ResourceVersion = ""

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			for err == nil {
				if ctx.Err() == context.DeadlineExceeded {
					return fmt.Errorf("pod %q failed to restart", p.Name)
				}
				_, err = pods.Get(context.TODO(), p.Name, metav1.GetOptions{})
				time.Sleep(time.Second)
			}

			if _, err := pods.Create(context.TODO(), &p, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("failed to create pod %q: %v", p.Name, err)
			}
		}
	}
	return nil
}

func disableAddonGCPAuth(cfg *config.ClusterConfig) error {
	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Clean up the files generated when enabling the addon
	creds := assets.NewMemoryAssetTarget([]byte{}, credentialsPath, readPermission)
	err := r.Remove(creds)
	if err != nil {
		return err
	}

	project := assets.NewMemoryAssetTarget([]byte{}, projectPath, readPermission)
	if err := r.Remove(project); err != nil {
		return err
	}

	client, err := service.K8s.GetCoreClient(cfg.Name)
	if err != nil {
		return err
	}

	namespaces, err := client.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// No need to check for an error here, if the secret doesn't exist, no harm done.
	for _, n := range namespaces.Items {
		if skipNamespace(n.Name) {
			continue
		}
		secrets := client.Secrets(n.Name)
		if err := secrets.Delete(context.TODO(), secretName, metav1.DeleteOptions{}); err != nil {
			klog.Infof("error deleting secret: %v", err)
		}

		serviceaccounts := client.ServiceAccounts(n.Name)
		salist, err := serviceaccounts.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.Infof("error getting service accounts: %v", err)
			return err
		}
		for _, sa := range salist.Items {
			for i, ps := range sa.ImagePullSecrets {
				if ps.Name == secretName {
					sa.ImagePullSecrets = append(sa.ImagePullSecrets[:i], sa.ImagePullSecrets[i+1:]...)
					if _, err := serviceaccounts.Update(context.TODO(), &sa, metav1.UpdateOptions{}); err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return nil
}

func verifyGCPAuthAddon(cc *config.ClusterConfig, name, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}

	// If we're in GCE and didn't actually start the gcp-auth pods, don't check for them.
	// We also don't want to actually set the addon as enabled, so just exit completely.
	if enable && !Force && detect.IsOnGCE() && os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return ErrSkipThisAddon
	}

	if Refresh {
		if err := refreshExistingPods(cc); err != nil {
			return err
		}
	}

	if err := verifyAddonStatusInternal(cc, name, val, "gcp-auth"); err != nil {
		return err
	}

	if enable && err == nil {
		out.Styled(style.Notice, "Your GCP credentials will now be mounted into every pod created in the {{.name}} cluster.", out.V{"name": cc.Name})
		out.Styled(style.Notice, "If you don't want your credentials mounted into a specific pod, add a label with the `gcp-auth-skip-secret` key to your pod configuration.")
		if !Refresh {
			out.Styled(style.Notice, "If you want existing pods to be mounted with credentials, either recreate them or rerun addons enable with --refresh.")
		}
	}

	return err
}

func skipNamespace(name string) bool {
	return name == metav1.NamespaceSystem || name == namespaceName
}
