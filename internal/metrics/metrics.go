package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// User Activity Metrics
	NewUsersTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_new_users_total",
		Help: "Total number of new user registrations.",
	})
	LoginAttemptsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_login_attempts_total",
		Help: "Total number of login attempts (successful and failed).",
	}, []string{"status"}) // status: "success" or "failed"

	// Application-Specific Feature Usage Metrics
	BookmarkCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_bookmark_created_total",
		Help: "Total number of bookmarks created.",
	})
	CategoryCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_category_created_total",
		Help: "Total number of categories created.",
	})
	CollectionCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_collection_created_total",
		Help: "Total number of collections created.",
	})
	TagCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_tag_created_total",
		Help: "Total number of tags created.",
	})
	SummaryGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_summary_generated_total",
		Help: "Total number of summaries generated.",
	})
	AISuggestionsGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_ai_suggestions_generated_total",
		Help: "Total number of AI suggestions generated.",
	})
)
