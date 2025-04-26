// file: buz/bus/bus_sub/sublist_test.go

package bus_sub

import (
	"log"
	"testing"

	"github.com/rskv-p/buzz/bus/bus_type"
)

//-------------------------------------------------
// TestMatchExactSubscription - Test for exact subject matching
//-------------------------------------------------

// TestMatchExactSubscription tests if exact subscriptions are matched correctly.
func TestMatchExactSubscription(t *testing.T) {
	sublist := NewSublist(10)

	// Create an exact subscription
	sub := &bus_type.Subscription{Subject: []byte("foo.bar.baz"), Queue: []byte("queue1"), Client: nil}

	// Insert the subscription
	sublist.Insert(sub)

	// Log all subscriptions after insertion
	log.Printf("All subscriptions after insertion:")
	for key, subs := range sublist.ExactCache {
		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}
	for key, subs := range sublist.WildcardCache {
		log.Printf("Wildcard Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}

	// Perform matching with the exact subject
	result := sublist.Match([]byte("foo.bar.baz"))

	// Log all subscriptions after matching
	log.Printf("All subscriptions after matching:")
	for key, subs := range sublist.ExactCache {
		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}
	for key, subs := range sublist.WildcardCache {
		log.Printf("Wildcard Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}

	// Check if the subscription is found
	if len(result.Psubs) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Psubs))
	}

	// Check if the exact subscription matched
	if string(result.Psubs[0].Subject) != "foo.bar.baz" {
		t.Fatal("expected subscription to match the exact pattern")
	}
}

//-------------------------------------------------
// TestRemoveSubscription - Test for removing a subscription
//-------------------------------------------------

// TestRemoveSubscription tests if subscriptions are removed correctly.
func TestRemoveSubscription(t *testing.T) {
	sublist := NewSublist(10)

	// Create an exact subscription
	sub := &bus_type.Subscription{Subject: []byte("foo.bar.baz"), Queue: []byte("queue1"), Client: nil}

	// Insert the subscription
	sublist.Insert(sub)

	// Log subscription after insertion
	log.Printf("All subscriptions after insertion:")
	for key, subs := range sublist.ExactCache {
		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}
	for key, subs := range sublist.WildcardCache {
		log.Printf("Wildcard Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}

	// Remove the subscription
	sublist.Remove(sub)

	// Log subscription after removal
	log.Printf("All subscriptions after removal:")
	for key, subs := range sublist.ExactCache {
		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}
	for key, subs := range sublist.WildcardCache {
		log.Printf("Wildcard Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}

	// Check that no subscription is found after removal
	result := sublist.Match([]byte("foo.bar.baz"))
	if len(result.Psubs) != 0 {
		t.Fatal("expected no matches after removal")
	}
}

//-------------------------------------------------
// MatchSubjectWithWildcard - Matching subscriptions with wildcard filter
//-------------------------------------------------

// MatchSubjectWithWildcard applies wildcard matching to subscriptions.
func (s *Sublist) MatchSubjectWithWildcard(filter []byte) *SublistResult {
	res := &SublistResult{}

	// Use tree.Match to find subscriptions matching the wildcard filter
	s.Tree.Match(filter, func(subject []byte, sub **bus_type.Subscription) {
		if sub != nil && *sub != nil {
			res.Psubs = append(res.Psubs, *sub) // Add the matched subscription to the result
		}
	})

	return res
}

//-------------------------------------------------
// TestMatchWildcardSubscription - Test for wildcard subject matching
//-------------------------------------------------

// TestMatchWildcardSubscription tests if subscriptions match correctly with wildcard filters.
func TestMatchWildcardSubscription(t *testing.T) {
	sublist := NewSublist(10)

	// Create exact subscriptions
	sub1 := &bus_type.Subscription{Subject: []byte("foo.bar.baz"), Queue: []byte("queue1"), Client: nil}
	sub2 := &bus_type.Subscription{Subject: []byte("foo.x.baz"), Queue: []byte("queue2"), Client: nil}
	sub3 := &bus_type.Subscription{Subject: []byte("foo.baz.qux"), Queue: []byte("queue3"), Client: nil}

	// Insert the subscriptions
	sublist.Insert(sub1)
	sublist.Insert(sub2)
	sublist.Insert(sub3)

	// Log subscriptions
	log.Printf("All subscriptions after insertion:")
	for key, subs := range sublist.ExactCache {
		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}

	// Apply matching with wildcard filter
	filter := []byte("foo.>") // Wildcard filter
	log.Printf("Matching with filter 'foo.>'...")
	result := sublist.MatchSubjectWithWildcard(filter)

	// Log results
	log.Printf("Matching result: %v", result.Psubs)

	// Check the number of matches
	if len(result.Psubs) != 3 {
		t.Fatalf("expected 3 matches for filter 'foo.>', got %d", len(result.Psubs))
	} else {
		log.Printf("Match for filter 'foo.>' passed with %d match(es)", len(result.Psubs))
	}

	// Verify that the correct subscriptions matched
	log.Printf("Verifying matched subscriptions...")

	expectedSubjects := []string{"foo.bar.baz", "foo.x.baz", "foo.baz.qux"}

	for _, expectedSubject := range expectedSubjects {
		found := false
		for _, sub := range result.Psubs {
			if string(sub.Subject) == expectedSubject {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected subscription with subject '%s', but not found", expectedSubject)
		} else {
			log.Printf("Correct subscription matched for subject: %s", expectedSubject)
		}
	}
}

//-------------------------------------------------
// TestHasInterest - Test for checking if there is interest in a subject
//-------------------------------------------------

// TestHasInterest tests if the bus has interest in a given subject.
func TestHasInterest(t *testing.T) {
	sublist := NewSublist(10)

	// Create a subscription
	sub := &bus_type.Subscription{Subject: []byte("foo.bar.baz"), Queue: []byte("queue1"), Client: nil}

	// Insert the subscription
	sublist.Insert(sub)

	// Check if there is interest in the given subject
	if !sublist.HasInterest([]byte("foo.bar.baz")) {
		t.Fatal("expected interest for subject foo.bar.baz")
	}

	// Check that there is no interest in a non-existing subject
	if sublist.HasInterest([]byte("foo.baz")) {
		t.Fatal("expected no interest for subject foo.baz")
	}
}

//-------------------------------------------------
// TestInsertMultipleSubscriptions - Test for inserting multiple subscriptions
//-------------------------------------------------

// TestInsertMultipleSubscriptions tests if multiple subscriptions are inserted correctly.
func TestInsertMultipleSubscriptions(t *testing.T) {
	sublist := NewSublist(10)

	// Create multiple subscriptions with the same and different subjects
	sub1 := &bus_type.Subscription{Subject: []byte("foo.bar.baz"), Queue: []byte("queue1"), Client: nil}
	sub2 := &bus_type.Subscription{Subject: []byte("foo.bar.baz"), Queue: []byte("queue2"), Client: nil}
	sub3 := &bus_type.Subscription{Subject: []byte("foo.baz.bar"), Queue: []byte("queue3"), Client: nil}

	// Insert the subscriptions
	sublist.Insert(sub1)
	sublist.Insert(sub2)
	sublist.Insert(sub3)

	// Log all subscriptions after insertion
	log.Printf("All subscriptions after insertion:")
	for key, subs := range sublist.ExactCache {
		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}
	for key, subs := range sublist.WildcardCache {
		log.Printf("Wildcard Cache: Subject = %s, Subscriptions = %+v", key, subs)
	}

	// Verify all subscriptions are added correctly
	result1 := sublist.Match([]byte("foo.bar.baz"))
	result2 := sublist.Match([]byte("foo.baz.bar"))

	if len(result1.Psubs) != 2 {
		t.Fatalf("expected 2 matches for subject foo.bar.baz, got %d", len(result1.Psubs))
	}

	if len(result2.Psubs) != 1 {
		t.Fatalf("expected 1 match for subject foo.baz.bar, got %d", len(result2.Psubs))
	}
}

//-------------------------------------------------
// TestCacheEviction - Test for cache eviction when cache limit is reached
//-------------------------------------------------

// TestCacheEviction tests if cache eviction works correctly when the cache limit is reached.
// func TestCacheEviction(t *testing.T) {
// 	sublist := NewSublist(3) // Set cache size to 3

// 	// Create subscriptions
// 	sub1 := &bus_type.Subscription{Subject: []byte("foo.bar.baz"), Queue: []byte("queue1"), Client: nil}
// 	sub2 := &bus_type.Subscription{Subject: []byte("foo.baz.bar"), Queue: []byte("queue2"), Client: nil}
// 	sub3 := &bus_type.Subscription{Subject: []byte("bar.foo.baz"), Queue: []byte("queue3"), Client: nil}
// 	sub4 := &bus_type.Subscription{Subject: []byte("baz.foo.bar"), Queue: []byte("queue4"), Client: nil}

// 	// Insert subscriptions
// 	sublist.Insert(sub1)
// 	sublist.Insert(sub2)
// 	sublist.Insert(sub3)

// 	// Log all subscriptions after insertion
// 	log.Printf("All subscriptions after insertion:")
// 	for key, subs := range sublist.ExactCache {
// 		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
// 	}

// 	// Insert one more subscription, which should cause eviction
// 	sublist.Insert(sub4)

// 	// Log all subscriptions after eviction
// 	log.Printf("All subscriptions after eviction:")
// 	for key, subs := range sublist.ExactCache {
// 		log.Printf("Exact Cache: Subject = %s, Subscriptions = %+v", key, subs)
// 	}

// 	// Check that there are only 3 items in the cache
// 	if len(sublist.ExactCache) != 3 {
// 		t.Fatalf("expected 3 items in cache, got %d", len(sublist.ExactCache))
// 	}

// 	// Check that sub1 was evicted, as the cache is full
// 	if _, found := sublist.ExactCache[string(sub1.Subject)]; found {
// 		t.Fatal("expected sub1 to be evicted from cache")
// 	} else {
// 		log.Printf("Correct eviction: sub1 (foo.bar.baz) was evicted from cache")
// 	}
// }
