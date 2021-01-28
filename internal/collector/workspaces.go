package collector

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/kaizendorks/terraform-cloud-exporter/internal/setup"

	tfe "github.com/hashicorp/go-tfe"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// workspaces is the Metric subsystem we use.
	workspaces = "workspaces"
)

// Metric descriptors.
var (
	WorkspacesCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, workspaces, "count"),
		"Total number of existing workspaces.",
		[]string{}, nil,
	)
	WorkspacesInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, workspaces, "workspace_info"),
		"Information about existing workspaces",
		[]string{"id", "name", "organization", "terraform_version", "created_at", "environment", "current_run", "current_run_status", "current_run_created_at"}, nil,
	)
)

// ScrapeWorkspaces scrapes metrics about the workspaces.
type ScrapeWorkspaces struct{}

func init() {
	Scrapers = append(Scrapers, ScrapeWorkspaces{})
}

// Name of the Scraper. Should be unique.
func (ScrapeWorkspaces) Name() string {
	return workspaces
}

// Help describes the role of the Scraper.
func (ScrapeWorkspaces) Help() string {
	return "Scrape information from the Workspaces API: https://www.terraform.io/docs/cloud/api/workspaces.html"
}

// Version of Terraform Cloud/Enterprise API from which scraper is available.
func (ScrapeWorkspaces) Version() string {
	return "v2"
}

// TODO: We might want to allow the user to control pageSize via cli/config
// This could be handy for users hitting API rate limits (30 per sec).
// Investigate performance of (100 requests for 1 item) vs (1 request for 100 items).
func getWorkspacesListPage(ctx context.Context, pageNumber int, config *setup.Config, ch chan<- prometheus.Metric) error {
	include := "current_run"
	workspacesList, err := config.Client.Workspaces.List(ctx, config.Organization, tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{
			PageSize:   20,
			PageNumber: pageNumber,
		},
		Include: &include,
	})
	if err != nil {
		return err
	}

	for _, w := range workspacesList.Items {
		select {
		case ch <- prometheus.MustNewConstMetric(
			WorkspacesInfo,
			prometheus.GaugeValue,
			1,
			w.ID,
			w.Name,
			w.Organization.Name,
			w.TerraformVersion,
			w.CreatedAt.String(),
			w.Environment,
			getCurrentRunID(w.CurrentRun),
			getCurrentRunStatus(w.CurrentRun),
			getCurrentRunCreatedAt(w.CurrentRun),
		):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeWorkspaces) Scrape(ctx context.Context, config *setup.Config, ch chan<- prometheus.Metric) error {
	// TODO: Dummy list call to get to get the number of workspaces.
	//       Investigate if there is a better way to get the workspace count.
	workspacesList, err := config.Client.Workspaces.List(ctx, config.Organization, tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{PageSize: 20},
	})
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(WorkspacesCount, prometheus.GaugeValue, float64(workspacesList.Pagination.TotalCount))

	g, ctx := errgroup.WithContext(ctx)
	for i := 1; i <= workspacesList.Pagination.TotalPages; i++ {
		pageNumber := i
		g.Go(func() error {
			return getWorkspacesListPage(ctx, pageNumber, config, ch)
		})
	}

	return g.Wait()
}

func getCurrentRunID(r *tfe.Run) string {
	if r == nil {
		return "na"
	}

	return r.ID
}

func getCurrentRunStatus(r *tfe.Run) string {
	if r == nil {
		return "na"
	}

	return string(r.Status)
}

func getCurrentRunCreatedAt(r *tfe.Run) string {
	if r == nil {
		return "na"
	}

	return r.CreatedAt.String()
}
