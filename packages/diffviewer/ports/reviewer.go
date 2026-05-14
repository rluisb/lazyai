package ports

import "github.com/rluisb/lazyai/packages/diffviewer/domain"

// Reviewer runs a prepared review and returns the user's final response.
type Reviewer interface {
	Run() (domain.ReviewResponse, error)
}
