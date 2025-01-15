/*
Copyright 2021 Flant JSC

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

package certificate

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudflare/cfssl/helpers"
	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsCertificateExpiringSoon(cert []byte, durationLeft time.Duration) (bool, error) {
	c, err := helpers.ParseCertificatePEM(cert)
	if err != nil {
		return false, fmt.Errorf("certificate cannot parsed: %v", err)
	}
	if time.Until(c.NotAfter) < durationLeft {
		return true, nil
	}
	return false, nil
}

// modified client-go@v0.29.8/util/certificate/csr/csr.go
//
// WaitForCertificate waits for a certificate to be issued until timeout, or returns an error.
func WaitForCertificate(ctx context.Context, clientWOWatch client.Client, reqName string, reqUID types.UID, logger pkg.Logger) (certData []byte, err error) {
	c, ok := clientWOWatch.(client.WithWatch)
	if !ok {
		return nil, errors.New("client without watch")
	}

	fieldSelector := client.MatchingFields{"metadata.name": reqName}

	var lw *cache.ListWatch
	var obj runtime.Object
	for {
		// see if the v1 API is available
		if err := c.List(ctx, &certificatesv1.CertificateSigningRequestList{}, fieldSelector); err == nil {
			// watch v1 objects
			obj = &certificatesv1.CertificateSigningRequest{}
			lw = &cache.ListWatch{
				ListFunc: func(_ metav1.ListOptions) (runtime.Object, error) {
					list := new(certificatesv1.CertificateSigningRequestList)
					err := c.List(ctx, list, fieldSelector)
					return list, err
				},
				WatchFunc: func(_ metav1.ListOptions) (watch.Interface, error) {
					w, err := c.Watch(ctx, new(certificatesv1.CertificateSigningRequestList), fieldSelector)
					return w, err
				},
			}

			break
		}

		logger.Info("error fetching v1 certificate signing request", log.Err(err))

		// return if we've timed out
		if err := ctx.Err(); err != nil {
			return nil, wait.ErrorInterrupted(err)
		}

		// see if the v1beta1 API is available
		if err := c.List(ctx, &certificatesv1beta1.CertificateSigningRequestList{}, fieldSelector); err == nil {
			// watch v1beta1 objects
			obj = &certificatesv1beta1.CertificateSigningRequest{}
			lw = &cache.ListWatch{
				ListFunc: func(_ metav1.ListOptions) (runtime.Object, error) {
					list := new(certificatesv1beta1.CertificateSigningRequestList)
					err := c.List(ctx, list, fieldSelector)
					return list, err
				},
				WatchFunc: func(_ metav1.ListOptions) (watch.Interface, error) {
					w, err := c.Watch(ctx, new(certificatesv1beta1.CertificateSigningRequestList), fieldSelector)
					return w, err
				},
			}

			break
		}

		logger.Info("error fetching v1beta1 certificate signing request", log.Err(err))

		// return if we've timed out
		if err := ctx.Err(); err != nil {
			return nil, wait.ErrorInterrupted(err)
		}

		// wait and try again
		time.Sleep(time.Second)
	}

	var issuedCertificate []byte
	_, err = watchtools.UntilWithSync(
		ctx,
		lw,
		obj,
		nil,
		func(event watch.Event) (bool, error) {
			switch event.Type {
			case watch.Modified, watch.Added:
			case watch.Deleted:
				return false, fmt.Errorf("csr %q was deleted", reqName)
			default:
				return false, nil
			}

			switch csr := event.Object.(type) {
			case *certificatesv1.CertificateSigningRequest:
				if csr.UID != reqUID {
					return false, fmt.Errorf("csr %q changed UIDs", csr.Name)
				}
				approved := false
				for _, c := range csr.Status.Conditions {
					if c.Type == certificatesv1.CertificateDenied {
						return false, fmt.Errorf("certificate signing request is denied, reason: %v, message: %v", c.Reason, c.Message)
					}
					if c.Type == certificatesv1.CertificateFailed {
						return false, fmt.Errorf("certificate signing request failed, reason: %v, message: %v", c.Reason, c.Message)
					}
					if c.Type == certificatesv1.CertificateApproved {
						approved = true
					}
				}
				if approved {
					if len(csr.Status.Certificate) > 0 {
						logger.Info("certificate signing request is issued", slog.String("request", csr.Name))
						issuedCertificate = csr.Status.Certificate
						return true, nil
					}
					logger.Info("certificate signing request is approved, waiting to be issued", slog.String("request", csr.Name))
				}

			case *certificatesv1beta1.CertificateSigningRequest:
				if csr.UID != reqUID {
					return false, fmt.Errorf("csr %q changed UIDs", csr.Name)
				}
				approved := false
				for _, c := range csr.Status.Conditions {
					if c.Type == certificatesv1beta1.CertificateDenied {
						return false, fmt.Errorf("certificate signing request is denied, reason: %v, message: %v", c.Reason, c.Message)
					}
					if c.Type == certificatesv1beta1.CertificateFailed {
						return false, fmt.Errorf("certificate signing request failed, reason: %v, message: %v", c.Reason, c.Message)
					}
					if c.Type == certificatesv1beta1.CertificateApproved {
						approved = true
					}
				}
				if approved {
					if len(csr.Status.Certificate) > 0 {
						logger.Info("certificate signing request is issued", slog.String("request", csr.Name))
						issuedCertificate = csr.Status.Certificate
						return true, nil
					}
					logger.Info("certificate signing request is approved, waiting to be issued", slog.String("request", csr.Name))
				}

			default:
				return false, fmt.Errorf("unexpected type received: %T", event.Object)
			}

			return false, nil
		},
	)
	if wait.Interrupted(err) {
		return nil, wait.ErrorInterrupted(err)
	}
	if err != nil {
		return nil, formatError("cannot watch on the certificate signing request: %v", err)
	}

	return issuedCertificate, nil
}

// formatError preserves the type of an API message but alters the message. Expects
// a single argument format string, and returns the wrapped error.
func formatError(format string, err error) error {
	if s, ok := err.(apierrors.APIStatus); ok {
		se := &apierrors.StatusError{ErrStatus: s.Status()}
		se.ErrStatus.Message = fmt.Sprintf(format, se.ErrStatus.Message)
		return se
	}
	return fmt.Errorf(format, err)
}
