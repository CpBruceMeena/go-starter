package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/go-starter/internal/business"
	"github.com/your-org/go-starter/internal/logger"
)

// ExampleCleanupTask is a sample task that cleans old data
func ExampleCleanupTask(ctx context.Context, log *logger.Logger) TaskFunc {
	return func(taskCtx context.Context) error {
		log.InfoContext(taskCtx, "running cleanup task",
			"operation", "delete_old_records",
		)

		// Simulate cleanup work
		time.Sleep(100 * time.Millisecond)

		// In real application:
		// err := repo.DeleteOlderThan(taskCtx, time.Now().AddDate(0, 0, -30))
		// if err != nil {
		//     return err
		// }

		log.InfoContext(taskCtx, "cleanup task completed",
			"operation", "delete_old_records",
			"records_deleted", 42,
		)

		return nil
	}
}

// ExampleSyncTask is a sample task that syncs data from external API
func ExampleSyncTask(ctx context.Context, log *logger.Logger, service business.UserService) TaskFunc {
	return func(taskCtx context.Context) error {
		log.InfoContext(taskCtx, "running sync task",
			"operation", "sync_external_data",
		)

		// In real application:
		// resp, err := externalAPI.GetUsers(taskCtx)
		// if err != nil {
		//     return fmt.Errorf("failed to fetch external users: %w", err)
		// }
		//
		// for _, user := range resp.Users {
		//     req := &models.CreateUserRequest{
		//         Email: user.Email,
		//         Name:  user.Name,
		//     }
		//     if _, err := service.CreateUser(taskCtx, req); err != nil {
		//         log.ErrorContext(taskCtx, "failed to create user", "error", err.Error())
		//     }
		// }

		log.InfoContext(taskCtx, "sync task completed",
			"operation", "sync_external_data",
			"users_synced", 10,
		)

		return nil
	}
}

// ExampleHealthCheckTask is a sample task that performs health checks
func ExampleHealthCheckTask(ctx context.Context, log *logger.Logger) TaskFunc {
	return func(taskCtx context.Context) error {
		log.InfoContext(taskCtx, "running health check task")

		// Check database connectivity
		// dbOk := checkDatabaseHealth(taskCtx)
		// cacheOk := checkCacheHealth(taskCtx)
		// externalOk := checkExternalAPIHealth(taskCtx)

		// For demo purposes
		dbOk := true
		cacheOk := true
		externalOk := true

		if !dbOk || !cacheOk || !externalOk {
			return fmt.Errorf("health check failed: db=%v, cache=%v, external=%v",
				dbOk, cacheOk, externalOk)
		}

		log.InfoContext(taskCtx, "health check completed",
			"database_ok", dbOk,
			"cache_ok", cacheOk,
			"external_ok", externalOk,
		)

		return nil
	}
}

// ExampleReportGenerationTask is a sample task that generates reports
func ExampleReportGenerationTask(ctx context.Context, log *logger.Logger) TaskFunc {
	return func(taskCtx context.Context) error {
		log.InfoContext(taskCtx, "starting report generation",
			"report_type", "daily_summary",
		)

		startTime := time.Now()

		// In real application:
		// stats := repo.GetDailyStats(taskCtx, time.Now())
		// report := generateReport(stats)
		// err := saveReport(taskCtx, report)
		// if err != nil {
		//     return fmt.Errorf("failed to save report: %w", err)
		// }

		duration := time.Since(startTime)

		log.InfoContext(taskCtx, "report generation completed",
			"report_type", "daily_summary",
			"duration_ms", duration.Milliseconds(),
			"recipients", 5,
		)

		return nil
	}
}

// ExampleNotificationTask is a sample task that sends notifications
func ExampleNotificationTask(ctx context.Context, log *logger.Logger) TaskFunc {
	return func(taskCtx context.Context) error {
		log.InfoContext(taskCtx, "processing notifications",
			"operation", "send_pending_notifications",
		)

		// In real application:
		// notifications := repo.GetPendingNotifications(taskCtx)
		// for _, notif := range notifications {
		//     err := emailService.Send(taskCtx, notif)
		//     if err != nil {
		//         log.ErrorContext(taskCtx, "failed to send notification", "error", err.Error())
		//     } else {
		//         repo.MarkNotificationSent(taskCtx, notif.ID)
		//     }
		// }

		log.InfoContext(taskCtx, "notifications processed",
			"operation", "send_pending_notifications",
			"sent", 3,
			"failed", 0,
		)

		return nil
	}
}

// RegisterExampleTasks registers all example tasks
func RegisterExampleTasks(w *Worker, log *logger.Logger, service business.UserService) {
	// Cleanup task - runs every hour
	w.RegisterTask(Task{
		Name:     "cleanup_old_data",
		Interval: 1 * time.Hour,
		Timeout:  5 * time.Minute,
		Fn:       ExampleCleanupTask(context.Background(), log),
	})

	// Sync task - runs every 30 minutes
	w.RegisterTask(Task{
		Name:     "sync_external_data",
		Interval: 30 * time.Minute,
		Timeout:  2 * time.Minute,
		Fn:       ExampleSyncTask(context.Background(), log, service),
	})

	// Health check - runs every 5 minutes
	w.RegisterTask(Task{
		Name:     "health_check",
		Interval: 5 * time.Minute,
		Timeout:  30 * time.Second,
		Fn:       ExampleHealthCheckTask(context.Background(), log),
	})

	// Report generation - runs daily at specific time
	w.RegisterTask(Task{
		Name:     "daily_report_generation",
		Interval: 24 * time.Hour,
		Timeout:  10 * time.Minute,
		Fn:       ExampleReportGenerationTask(context.Background(), log),
	})

	// Notification processing - runs every 5 minutes
	w.RegisterTask(Task{
		Name:     "process_notifications",
		Interval: 5 * time.Minute,
		Timeout:  1 * time.Minute,
		Fn:       ExampleNotificationTask(context.Background(), log),
	})
}
