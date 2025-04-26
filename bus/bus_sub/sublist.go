// file: buz/bus/bus_sub/sublist.go

package bus_sub

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/rskv-p/buzz/bus/bus_comm"
	"github.com/rskv-p/buzz/bus/bus_type"
	"github.com/rskv-p/subtree"
)

//-------------------------------------------------
// Sublist - Manages subscriptions
//-------------------------------------------------

// Sublist manages subscriptions, including exact and wildcard matches.
type Sublist struct {
	Tree          *subtree.SubjectTree[*bus_type.Subscription] // Tree structure for subject subscriptions
	ExactCache    map[string][]*bus_type.Subscription          // Cache for exact matches
	WildcardCache map[string][]*bus_type.Subscription          // Cache for wildcard matches
	cacheSize     int                                          // Cache size limit
}

//-------------------------------------------------
// NewSublist - Creates a new Sublist
//-------------------------------------------------

// NewSublist creates a new Sublist with a specified cache size.
func NewSublist(cacheSize int) *Sublist {
	bus_comm.Infof("Creating a new Sublist with cache size: %d", cacheSize)
	return &Sublist{
		Tree:          subtree.NewSubjectTree[*bus_type.Subscription](),
		ExactCache:    make(map[string][]*bus_type.Subscription),
		WildcardCache: make(map[string][]*bus_type.Subscription),
		cacheSize:     cacheSize,
	}
}

//-------------------------------------------------
// Insert - Adds a Subscription
//-------------------------------------------------

// Insert adds a Subscription to the tree and updates the cache.
func (s *Sublist) Insert(sub *bus_type.Subscription) {
	bus_comm.Infof("Inserting subscription: subject = %s", string(sub.Subject))
	s.Tree.Insert(sub.Subject, sub)
	key := string(sub.Subject)

	// Check if it's a wildcard and update the appropriate cache
	if isWildcard(key) {
		s.WildcardCache[key] = append(s.WildcardCache[key], sub)
		bus_comm.Infof("Inserted wildcard Subscription: subject = %s", key)
	} else {
		s.ExactCache[key] = append(s.ExactCache[key], sub)

		// Evict cache entry if it exceeds the cache size limit
		if len(s.ExactCache) > s.cacheSize {
			for k := range s.ExactCache {
				delete(s.ExactCache, k)
				bus_comm.Infof("Evicted exact Subscription from cache: subject = %s", k)
				break
			}
		}

		bus_comm.Infof("Inserted exact Subscription: subject = %s", key)
	}
}

//-------------------------------------------------
// Remove - Deletes a Subscription
//-------------------------------------------------

// Remove deletes a Subscription from the tree.
func (s *Sublist) Remove(sub *bus_type.Subscription) error {
	key := string(sub.Subject)
	bus_comm.Infof("Removing subscription: subject = %s", key)

	// Check if the subscription exists in the tree
	existingSub, exists := s.Tree.Find(sub.Subject)
	if !exists || existingSub == nil {
		bus_comm.Errorf("Subscription with subject %s not found in tree", key)
		return fmt.Errorf("subscription with subject %s not found in tree", key)
	}

	// Try to remove the subscription from the tree
	s.Tree.Delete(sub.Subject)

	// Remove from the appropriate cache (exact or wildcard)
	if isWildcard(key) {
		if _, exists := s.WildcardCache[key]; !exists {
			bus_comm.Errorf("Wildcard subscription with subject %s not found in cache", key)
			return fmt.Errorf("wildcard subscription with subject %s not found in cache", key)
		}
		delete(s.WildcardCache, key)
		bus_comm.Infof("Removed wildcard Subscription: subject = %s", key)
	} else {
		if _, exists := s.ExactCache[key]; !exists {
			bus_comm.Errorf("Exact subscription with subject %s not found in cache", key)
			return fmt.Errorf("exact subscription with subject %s not found in cache", key)
		}
		delete(s.ExactCache, key)
		bus_comm.Infof("Removed exact Subscription: subject = %s", key)
	}

	return nil
}

//-------------------------------------------------
// SublistResult - Holds the subscriptions that match a subject
//-------------------------------------------------

// SublistResult holds the subscriptions that match a subject.
type SublistResult struct {
	Psubs []*bus_type.Subscription // List of matching subscriptions
}

//-------------------------------------------------
// Match - Returns Subscriptions matching the given subject
//-------------------------------------------------

// Match returns Subscriptions matching the given subject.
func (s *Sublist) Match(subject []byte) *SublistResult {
	bus_comm.Infof("Matching subscriptions for subject: %s", string(subject))
	res := &SublistResult{}

	// Debugging the current subscriptions
	bus_comm.Debugf("Exact cache: %v", s.ExactCache)
	bus_comm.Debugf("Wildcard cache: %v", s.WildcardCache)

	// Handle regular subjects
	if subs, ok := s.ExactCache[string(subject)]; ok {
		bus_comm.Infof("Exact match found for subject: %s", string(subject))
		res.Psubs = append(res.Psubs, subs...)
	}

	// Use the tree to match subscriptions (e.g., with wildcards)
	s.Tree.Match(subject, func(_ []byte, sub **bus_type.Subscription) {
		if sub != nil && *sub != nil && !contains(res.Psubs, *sub) {
			res.Psubs = append(res.Psubs, *sub)
			bus_comm.Infof("Matching subscription added: subject = %s", string(subject))
		}
	})

	bus_comm.Infof("Tree matched Subscriptions: count = %d, subject = %s", len(res.Psubs), string(subject))
	return res
}

//-------------------------------------------------
// HasInterest - Checks if any subscription matches the subject
//-------------------------------------------------

// HasInterest checks if any subscription matches the subject.
func (s *Sublist) HasInterest(subject []byte) bool {
	bus_comm.Infof("Checking interest for subject: %s", string(subject))
	found := false
	s.Tree.Match(subject, func(_ []byte, _ **bus_type.Subscription) {
		found = true
	})
	bus_comm.Infof("HasInterest: subject = %s, found = %v", string(subject), found)
	return found
}

//-------------------------------------------------
// Helpers - Helper Functions
//-------------------------------------------------

// isWildcard checks if the subject contains wildcard characters ('*' or '>').
func isWildcard(subj string) bool {
	return bytes.ContainsAny([]byte(subj), ">*/")
}

// contains checks if a subscription is already in the list.
func contains(list []*bus_type.Subscription, target *bus_type.Subscription) bool {
	for _, s := range list {
		if s == target {
			bus_comm.Debugf("Subscription already in list: %v", s)
			return true
		}
	}
	bus_comm.Debugf("Subscription not in list: %v", target)
	return false
}

//-------------------------------------------------
// Subject Transformation - Handling subject transformations
//-------------------------------------------------

// SubjectTransform handles the transformation of subjects based on specific rules.
type SubjectTransform struct {
	srcTokens  []string
	destTokens []string
}

//-------------------------------------------------
// NewSubjectTransform - Creates a new transformer for converting src → dest
//-------------------------------------------------

// NewSubjectTransform creates a new transformer for converting src → dest.
func NewSubjectTransform(src, dest string) (*SubjectTransform, error) {
	bus_comm.Infof("Creating subject transform: src = %s, dest = %s", src, dest)
	srcTokens := strings.Split(src, ".")
	destTokens := strings.Split(dest, ".")

	// Check if the number of wildcards matches between source and destination
	if countWildcards(srcTokens) != countWildcards(destTokens) {
		bus_comm.Errorf("Wildcard count mismatch in transform: src = %s, dest = %s", src, dest)
		return nil, errors.New("wildcard count mismatch between src and dest")
	}

	return &SubjectTransform{
		srcTokens:  srcTokens,
		destTokens: destTokens,
	}, nil
}

//-------------------------------------------------
// TransformSubject - Applies the transformation to the subject based on the src → dest rule
//-------------------------------------------------

// TransformSubject applies the transformation to the subject based on the src → dest rule.
func (st *SubjectTransform) TransformSubject(subject string) (string, error) {
	bus_comm.Infof("Transforming subject: %s", subject)
	inputTokens := strings.Split(subject, ".")
	mapping := make([]string, 0)
	i := 0

	// Apply the transformation rules to the source subject
	for _, token := range st.srcTokens {
		switch token {
		case "*":
			if i >= len(inputTokens) {
				bus_comm.Errorf("Subject too short for *: subject = %s", subject)
				return "", errors.New("subject too short for *")
			}
			mapping = append(mapping, inputTokens[i])
			i++
		case ">":
			if i >= len(inputTokens) {
				bus_comm.Errorf("No tokens available for >: subject = %s", subject)
				return "", errors.New("no tokens available for >")
			}
			mapping = append(mapping, strings.Join(inputTokens[i:], "."))
			i = len(inputTokens) // Move the index to the end
		default:
			if i >= len(inputTokens) || token != inputTokens[i] {
				bus_comm.Errorf("Subject does not match pattern: expected = %s, got = %s", token, inputTokens[i])
				return "", errors.New("subject does not match source pattern")
			}
			mapping = append(mapping, inputTokens[i])
			i++
		}
	}

	// Build the transformed result based on destTokens
	var result []string
	wcIndex := 0
	for _, token := range st.destTokens {
		if token == "*" || token == ">" {
			if wcIndex >= len(mapping) {
				bus_comm.Errorf("Not enough wildcards for destination: subject = %s", subject)
				return "", errors.New("not enough wildcard values to fill destination")
			}
			result = append(result, mapping[wcIndex])
			wcIndex++
		} else {
			result = append(result, token)
		}
	}

	// Join the final transformed subject
	transformed := strings.Join(result, ".")
	bus_comm.Infof("Transformed subject: input = %s, output = %s", subject, transformed)
	return transformed, nil
}

//-------------------------------------------------
// Helpers - Helper Functions
//-------------------------------------------------

// countWildcards counts the number of wildcard characters ('*' or '>') in the tokens.
func countWildcards(tokens []string) int {
	count := 0
	for _, t := range tokens {
		if t == "*" || t == ">" {
			count++
		}
	}
	return count
}

//-------------------------------------------------
// MatchWithQueue - Checks for subscriptions matching both subject and queue
//-------------------------------------------------

// MatchWithQueue checks for subscriptions matching both subject and queue.
func (s *Sublist) MatchWithQueue(subject []byte, queue []byte) *SublistResult {
	bus_comm.Infof("Matching subscriptions for subject: %s and queue: %s", string(subject), string(queue))
	res := &SublistResult{}

	// Handle exact subject matches and queue filtering
	if subs, ok := s.ExactCache[string(subject)]; ok {
		for _, sub := range subs {
			if string(sub.Queue) == string(queue) {
				res.Psubs = append(res.Psubs, sub)
			}
		}
	}

	// Check wildcard subscriptions as well
	s.Tree.Match(subject, func(_ []byte, sub **bus_type.Subscription) {
		if sub != nil && *sub != nil && string((*sub).Queue) == string(queue) {
			res.Psubs = append(res.Psubs, *sub)
		}
	})

	bus_comm.Infof("Matched %d subscriptions for subject %s and queue %s", len(res.Psubs), string(subject), string(queue))
	return res
}
