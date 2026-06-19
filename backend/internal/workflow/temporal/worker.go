package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// NewWorker creates a Temporal worker on the given task queue and registers
// all workflows and activities via the Activities struct.
func NewWorker(c client.Client, taskQueue string, activities *Activities) worker.Worker {
	w := worker.New(c, taskQueue, worker.Options{})

	// --- Workflows ---
	w.RegisterWorkflow(ServiceCreateWorkflow)
	w.RegisterWorkflow(ReleaseWorkflow)
	w.RegisterWorkflow(RollbackWorkflow)
	w.RegisterWorkflow(DomainBindWorkflow)
	w.RegisterWorkflow(CertificateIssueWorkflow)
	w.RegisterWorkflow(LogAttachWorkflow)

	// --- Service creation activities ---
	w.RegisterActivity(activities.CreateServiceRecord)
	w.RegisterActivity(activities.CreateRepo)
	w.RegisterActivity(activities.GenerateCIConfig)
	w.RegisterActivity(activities.CreateTestEnvironment)
	w.RegisterActivity(activities.BindDomain)
	w.RegisterActivity(activities.RequestCertificate)
	w.RegisterActivity(activities.AttachLogSource)
	w.RegisterActivity(activities.AttachMetricPanel)
	w.RegisterActivity(activities.FinalizeService)

	// --- Release activities ---
	w.RegisterActivity(activities.ValidateReleasePrecondition)
	w.RegisterActivity(activities.RunRiskAnalysis)
	w.RegisterActivity(activities.CreateBuild)
	w.RegisterActivity(activities.WaitForBuildComplete)
	w.RegisterActivity(activities.Deploy)
	w.RegisterActivity(activities.HealthCheck)
	w.RegisterActivity(activities.ObserveWindow)
	w.RegisterActivity(activities.AutoRollback)
	w.RegisterActivity(activities.FinalizeRelease)

	// --- Rollback activities ---
	w.RegisterActivity(activities.ValidateRollback)
	w.RegisterActivity(activities.GetPreviousVersion)
	w.RegisterActivity(activities.ExecuteRollback)
	w.RegisterActivity(activities.VerifyRollback)
	w.RegisterActivity(activities.CreateRollbackRecord)

	// --- Domain binding activities ---
	w.RegisterActivity(activities.CreateDNSEntry)
	w.RegisterActivity(activities.WaitForDNSPropagation)
	w.RegisterActivity(activities.WaitCertReady)

	// --- Certificate storage activities ---
	w.RegisterActivity(activities.StoreCertSecret)

	// --- Log attachment activities ---
	w.RegisterActivity(activities.CreateLogstore)
	w.RegisterActivity(activities.ConfigureLogAgent)
	w.RegisterActivity(activities.CreateLogDashboard)
	w.RegisterActivity(activities.VerifyLogIngestion)

	return w
}

// Start runs the worker, blocking until interrupted.
func Start(w worker.Worker) error {
	return w.Run(worker.InterruptCh())
}

// Stop signals the worker to stop.
func Stop(w worker.Worker) {
	if w != nil {
		w.Stop()
	}
}
