package repository

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/mqufflc/whodidthechores/internal/repository/postgres"
)

type TaskReport struct {
	User  User  `json:"user"`
	Chore Chore `json:"chore"`
	Sum   int64 `json:"sum"`
}

type SingleChoreReport struct {
	Chore string `json:"chore"`
	Sum   int64  `json:"sum"`
}

type SingleUserReport struct {
	User string `json:"user"`
	Sum  int64  `json:"sum"`
}

type Report struct {
	Report map[string]map[string]int64 `json:"report"`
	Users  []string                    `json:"users"`
	Chores []string                    `json:"chores"`
}

func GenerateUserReport(tasks []TaskReport) map[string][]SingleChoreReport {
	report := make(map[string][]SingleChoreReport)
	for _, task := range tasks {
		newReport := SingleChoreReport{Chore: task.Chore.Name, Sum: task.Sum}
		existingReport := report[task.User.Name]
		report[task.User.Name] = append(existingReport, newReport)
	}
	return report
}

func GenerateReport(tasks []TaskReport) Report {
	report := make(map[string]map[string]int64)
	var users []string
	var chores []string
	for _, task := range tasks {
		if !slices.Contains(users, task.User.Name) {
			users = append(users, task.User.Name)
		}
		if !slices.Contains(chores, task.Chore.Name) {
			chores = append(chores, task.Chore.Name)
		}
		existingReport, ok := report[task.Chore.Name]
		if !ok {
			existingReport = make(map[string]int64)
		}
		existingReport[task.User.Name] = task.Sum
		report[task.Chore.Name] = existingReport
	}
	slices.SortFunc(users, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	slices.SortFunc(chores, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return Report{
		Report: report,
		Users:  users,
		Chores: chores,
	}
}

func (r *Repository) GetChoreReport(ctx context.Context, start time.Time, end time.Time) (Report, error) {
	reports, err := r.q.TasksReport(ctx, postgres.TasksReportParams{NotBefore: start, NotAfter: end})
	if err != nil {
		if sqlErr := taskPgError(err); sqlErr != nil {
			return Report{}, sqlErr
		}
		return Report{}, err
	}
	choreReports := make([]TaskReport, len(reports))
	for index, report := range reports {
		choreReports[index] = TaskReport{
			User:  User(report.User),
			Chore: Chore(report.Chore),
			Sum:   report.Sum,
		}
	}
	return GenerateReport(choreReports), nil
}
