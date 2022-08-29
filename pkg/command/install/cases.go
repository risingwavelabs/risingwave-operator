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

package install

import (
	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	IssuerName      = "crypt-test"
	CertificateName = "default.test.cert"
	SecretName      = "example-issuer-account-key"
)

var nginx = "nginx"

// This case is to test the cert-manager.
// After the test, the ClusterIssuer will be deleted.
var issuer = v1.ClusterIssuer{
	ObjectMeta: metav1.ObjectMeta{
		Name: IssuerName,
	},
	Spec: v1.IssuerSpec{
		IssuerConfig: v1.IssuerConfig{
			ACME: &cmacme.ACMEIssuer{
				Email:  "user@example.com",
				Server: "https://acme-staging-v02.api.letsencrypt.org/directory",
				PrivateKey: cmmeta.SecretKeySelector{
					LocalObjectReference: cmmeta.LocalObjectReference{
						Name: SecretName,
					},
				},
				Solvers: []cmacme.ACMEChallengeSolver{
					{
						HTTP01: &cmacme.ACMEChallengeSolverHTTP01{
							Ingress: &cmacme.ACMEChallengeSolverHTTP01Ingress{
								Class: &nginx,
							},
						},
					},
				},
			},
		},
	},
}

// This case is to test the cert-manager.
// After the test, the Certificate will be deleted.
var cf = v1.Certificate{
	ObjectMeta: metav1.ObjectMeta{
		Name:      CertificateName,
		Namespace: "default",
	},
	Spec: v1.CertificateSpec{
		DNSNames: []string{
			"default.test.cert",
		},
		SecretName: SecretName,
		IssuerRef: cmmeta.ObjectReference{
			Name: IssuerName,
			Kind: "ClusterIssuer",
		},
	},
}
