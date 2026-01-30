/*
 * Copyright 2023 RisingWave Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhook

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

// SetupWebhooksWithManager set up the webhooks.
func SetupWebhooksWithManager(mgr ctrl.Manager, openKruiseAvailable bool) error {
	if err := ctrl.NewWebhookManagedBy(mgr, &risingwavev1alpha1.RisingWave{}).
		WithCustomDefaulter(NewRisingWaveMutatingWebhook()).
		WithCustomValidator(NewRisingWaveValidatingWebhook(openKruiseAvailable)).
		Complete(); err != nil {
		return fmt.Errorf("unable to setup webhooks for risingwave: %w", err)
	}

	if err := ctrl.NewWebhookManagedBy(mgr, &risingwavev1alpha1.RisingWaveScaleView{}).
		WithCustomDefaulter(NewRisingWaveScaleViewMutatingWebhook(mgr.GetAPIReader())).
		WithCustomValidator(NewRisingWaveScaleViewValidatingWebhook(mgr.GetClient())).
		Complete(); err != nil {
		return fmt.Errorf("unable to setup webhooks for risingwave scale view: %w", err)
	}

	return nil
}
