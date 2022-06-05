/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package translate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Xuanwo/go-locale"
	"golang.org/x/text/language"

	"k8s.io/klog/v2"
	"k8s.io/minikube/translations"
)

var (
	// preferredLanguage is the default language messages will be output in
	preferredLanguage = language.AmericanEnglish
	// our default language
	defaultLanguage = language.AmericanEnglish

	// Translations is a translation map from strings that can be output to console
	// to its translation in the user's system locale.
	Translations map[string]interface{}
)

// T translates the given string to the preferred language.
func T(s string) string {
	if preferredLanguage == defaultLanguage {
		return s
	}

	if len(Translations) == 0 {
		return s
	}

	if t, ok := Translations[s]; ok {
		if len(t.(string)) > 0 && t.(string) != " " {
			return t.(string)
		}
	}

	return s
}

// DetermineLocale finds the system locale and sets the preferred language for output appropriately.
func DetermineLocale() {
	tag, err := locale.Detect()
	if err != nil {
		klog.V(1).Infof("Getting system locale failed: %v", err)
	}
	SetPreferredLanguage(tag)

	// Load translations for preferred language into memory.
	p := preferredLanguage.String()
	t, err := translations.Translations.ReadFile(fmt.Sprintf("%s.json", p))
	if err != nil {
		// Attempt to find a more broad locale, e.g. fr instead of fr-FR.
		if strings.Contains(p, "-") {
			p = strings.Split(p, "-")[0]
			t, err = translations.Translations.ReadFile(fmt.Sprintf("%s.json", p))
			if err != nil {
				klog.V(1).Infof("Failed to load translation file for %s: %v", p, err)
				return
			}
		} else {
			klog.V(1).Infof("Failed to load translation file for %s: %v", preferredLanguage.String(), err)
			return
		}
	}

	err = json.Unmarshal(t, &Translations)
	if err != nil {
		klog.V(1).Infof("Failed to populate translation map: %v", err)
	}

}

// SetPreferredLanguage configures which language future messages should use.
func SetPreferredLanguage(tag language.Tag) {
	// output message only if verbosity level is set and we still haven't got all the flags parsed in main()
	klog.V(1).Infof("Setting Language to %s ...", tag)
	preferredLanguage = tag
}

// GetPreferredLanguage returns the preferred language tag.
func GetPreferredLanguage() language.Tag {
	return preferredLanguage
}
