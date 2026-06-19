package temporal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// CertificateIssueResult is the workflow return value.
type CertificateIssueResult struct {
	CertID     int64  `json:"certId"`
	SecretRef  string `json:"secretRef"`
	Status     string `json:"status"`
}

// CertificateIssueWorkflow orchestrates certificate request, wait, and storage.
func CertificateIssueWorkflow(ctx workflow.Context, input RequestCertInput) (CertificateIssueResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CertificateIssueWorkflow started", "domain", input.DomainFQDN)

	result := CertificateIssueResult{}

	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         ExternalRetry,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	// 1. Request certificate
	var certOut RequestCertOutput
	if err := workflow.ExecuteActivity(ctx, ActivityRequestCert, input).Get(ctx, &certOut); err != nil {
		return result, err
	}
	result.CertID = certOut.CertID

	// 2. Wait for cert ready
	var waitOut WaitCertOutput
	if err := workflow.ExecuteActivity(ctx, ActivityWaitCert, WaitCertInput{
		CertID: certOut.CertID,
	}).Get(ctx, &waitOut); err != nil {
		return result, err
	}

	// 3. Store cert secret
	var storeOut StoreCertOutput
	if err := workflow.ExecuteActivity(ctx, ActivityStoreCert, StoreCertInput{
		CertID: certOut.CertID,
		Domain: input.DomainFQDN,
	}).Get(ctx, &storeOut); err != nil {
		return result, err
	}

	result.SecretRef = storeOut.SecretRef
	result.Status = "issued"
	return result, nil
}
