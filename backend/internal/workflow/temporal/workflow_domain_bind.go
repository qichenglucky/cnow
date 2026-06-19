package temporal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// DomainBindResult is the workflow return value.
type DomainBindResult struct {
	DomainFQDN string `json:"domainFqdn"`
	CertID     int64  `json:"certId"`
	Status     string `json:"status"`
}

// DomainBindWorkflow orchestrates DNS creation, certificate provisioning, and verification.
func DomainBindWorkflow(ctx workflow.Context, input BindDomainInput) (DomainBindResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("DomainBindWorkflow started", "serviceId", input.ServiceID, "envId", input.EnvID)

	result := DomainBindResult{}

	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         ExternalRetry,
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	// 1. Create DNS entry
	var dnsOut CreateDNSOutput
	domain := "svc.cnow.io" // placeholder
	if err := workflow.ExecuteActivity(ctx, ActivityCreateDNS, CreateDNSInput{
		Domain: domain,
	}).Get(ctx, &dnsOut); err != nil {
		return result, err
	}

	// 2. Wait for DNS propagation
	var waitDNSOut WaitDNSOutput
	if err := workflow.ExecuteActivity(ctx, ActivityWaitDNS, WaitDNSInput{
		Domain:  domain,
		EntryID: dnsOut.EntryID,
	}).Get(ctx, &waitDNSOut); err != nil {
		return result, err
	}

	// 3. Request certificate
	var certOut RequestCertOutput
	if err := workflow.ExecuteActivity(ctx, ActivityRequestCert, RequestCertInput{
		DomainFQDN: domain,
	}).Get(ctx, &certOut); err != nil {
		return result, err
	}

	// 4. Wait for cert ready
	var waitCertOut WaitCertOutput
	if err := workflow.ExecuteActivity(ctx, ActivityWaitCert, WaitCertInput{
		CertID: certOut.CertID,
	}).Get(ctx, &waitCertOut); err != nil {
		return result, err
	}

	// 5. Store cert secret
	var storeOut StoreCertOutput
	if err := workflow.ExecuteActivity(ctx, ActivityStoreCert, StoreCertInput{
		CertID: certOut.CertID,
		Domain: domain,
	}).Get(ctx, &storeOut); err != nil {
		return result, err
	}

	result.DomainFQDN = domain
	result.CertID = certOut.CertID
	result.Status = "bound"
	_ = storeOut
	return result, nil
}
