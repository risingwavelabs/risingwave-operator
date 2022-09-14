/*
 * Copyright 2022 Singularity Data
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

package object

import risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"

type ScaleViewLockManager struct {
	mgr *RisingWaveManager
}

func (svl *ScaleViewLockManager) GetScaleViewLock(v *risingwavev1alpha1.RisingWaveScaleView) *risingwavev1alpha1.RisingWaveScaleViewLock {
	scaleViews := svl.mgr.RisingWaveReader.RisingWave().Status.ScaleViews
	for _, sv := range scaleViews {
		if sv.Name == v.Name && sv.UID == v.UID {
			return &sv
		}
	}
	return nil
}

func (svl *ScaleViewLockManager) IsScaleViewLocked(v *risingwavev1alpha1.RisingWaveScaleView) bool {
	return svl.GetScaleViewLock(v) != nil
}

func (svl *ScaleViewLockManager) GrabScaleViewLockFor(v *risingwavev1alpha1.RisingWaveScaleView) error {
	// TODO implement me
	panic("implement me")
}
