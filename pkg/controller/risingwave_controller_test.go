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

package risingwave

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/ctrlkit"
	"github.com/singularity-data/risingwave-operator/pkg/options"
)

var schemeForTest = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(schemeForTest)
	_ = risingwavev1alpha1.AddToScheme(schemeForTest)

	opt := &options.InnerRisingWaveOptions{}
	lo.Must0(opt.BuildConfigFromFile("../options/test_config.yaml"))
}

type resultErr struct {
	reconcile.Result
	error
}

func (r *resultErr) Equals(result reconcile.Result, err error) bool {
	return r.Result == result && errors.Is(err, r.error)
}

func newResultErr(r reconcile.Result, err error) resultErr {
	return resultErr{Result: r, error: err}
}

type actionAssertionHook struct {
	t         *testing.T
	asserts   map[string]resultErr
	mustCover bool
}

func (h *actionAssertionHook) PostRun(ctx context.Context, logger logr.Logger, action string, result reconcile.Result, err error) {
	resultErr, ok := h.asserts[action]
	if !ok {
		return
	}

	if !resultErr.Equals(result, err) {
		fmt.Printf("%s\t[%s]\n", color.RedString("FAIL"), action)
		h.t.Fatalf("unexpected result and error: %v,%v, expected: %v,%v, action: %s", result, err, resultErr.Result, resultErr.error, action)
	} else {
		fmt.Printf("%s\t[%s]\n", color.GreenString("PASS"), action)
	}
}

func (h *actionAssertionHook) PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]client.Object) {
	if _, ok := h.asserts[action]; !ok {
		if h.mustCover {
			h.t.Fatalf("unexpected action: %s", action)
		}
	}
}

func newActionAsserts(t *testing.T, asserts map[string]resultErr, mustCover bool) ctrlkit.ActionHook {
	return &actionAssertionHook{
		t:         t,
		asserts:   asserts,
		mustCover: mustCover,
	}
}

func Test_RisingWaveController_New(t *testing.T) {
	risingwave := &risingwavev1alpha1.RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: risingwavev1alpha1.RisingWaveSpec{
			Arch: "arm64",
			MetaNode: &risingwavev1alpha1.MetaNodeSpec{
				Storage: &risingwavev1alpha1.MetaStorage{
					Type: risingwavev1alpha1.InMemory,
				},
			},
			ObjectStorage: &risingwavev1alpha1.ObjectStorageSpec{
				Memory: true,
			},
		},
		Status: risingwavev1alpha1.RisingWaveStatus{},
	}
	risingwave.Default()

	controller := &RisingWaveController{
		Client: fake.NewClientBuilder().
			WithScheme(schemeForTest).
			WithObjects(risingwave).
			Build(),
		ActionHookFactory: func() ctrlkit.ActionHook {
			return newActionAsserts(t, map[string]resultErr{
				// New => Initializing(true), Running(false)
				RisingWaveAction_BarrierFirstTimeObserved:        newResultErr(ctrlkit.Continue()),
				RisingWaveAction_MarkConditionInitializingAsTrue: newResultErr(ctrlkit.Continue()),
				RisingWaveAction_MarkConditionRunningAsFalse:     newResultErr(ctrlkit.Continue()),

				// Initializing(true) => stop
				RisingWaveAction_BarrierConditionInitializingIsTrue: newResultErr(ctrlkit.Exit()),

				// Running(true) => stop
				RisingWaveAction_BarrierConditionRunningIsTrue: newResultErr(ctrlkit.Exit()),

				// Upgrading(true) => stop
				RisingWaveAction_BarrierConditionUpgradingIsTrue: newResultErr(ctrlkit.Exit()),

				// Sync status with running stats
				RisingWaveAction_CollectRunningStatisticsAndSyncStatus: newResultErr(ctrlkit.Continue()),

				// Update status
				RisingWaveAction_UpdateRisingWaveStatusViaClient: newResultErr(ctrlkit.Continue()),
			}, true)
		},
	}

	logger := zap.New(zap.UseDevMode(true))
	_, err := controller.Reconcile(log.IntoContext(context.Background(), logger), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "example",
			Namespace: "default",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
