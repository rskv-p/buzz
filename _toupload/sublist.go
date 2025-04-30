// file: buzz/mod/m_bus/bus_sub/sublist.go

package bus_sub

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/rskv-p/buzz/pkg/x_log"
	"github.com/rskv-p/buzz/typ"
	"github.com/rskv-p/subtree"
)

//-------------------------------------------------
// Sublist - Manages subscriptions
//-------------------------------------------------

// Sublist manages subscriptions, including exact and wildcard matches.
type Sublist struct {
	Tree          *subtree.SubjectTree[*typ.Subscription] // Tree structure for subject subscriptions
	ExactCache    map[string][]*typ.Subscription          // Cache for exact matches
	WildcardCache map[string][]*typ.Subscription          // Cache for wildcard matches
	cacheSize     int                                     // Cache size limit
}

//-------------------------------------------------
// NewSublist - Creates a new Sublist
//-------------------------------------------------

// NewSublist creates a new Sublist with a specified cache size.
func NewSublist(cacheSize int) *Sublist {
	x_log.Info("Creating a new Sublist with cache size:", cacheSize)
	return &Sublist{
		Tree:          subtree.NewSubjectTree[*typ.Subscription](),
		ExactCache:    make(map[string][]*typ.Subscription),
		WildcardCache: make(map[string][]*typ.Subscription),
		cacheSize:     cacheSize,
	}
}

//-------------------------------------------------
// Insert - Adds a Subscription
//-------------------------------------------------

// Insert adds a Subscription to the tree and updates the cache.
func (s *Sublist) Insert(sub *typ.Subscription) {
	x_log.Info("Inserting subscription: subject =", string(sub.Subject))
	s.Tree.Insert(sub.Subject, sub)
	key := string(sub.Subject)

	// Check if it's a wildcard and update the appropriate cache
	if isWildcard(key) {
		s.WildcardCache[key] = append(s.WildcardCache[key], sub)
		x_log.Info("Inserted wildcard Subscription: subject =", key)
	} else {
		s.ExactCache[key] = append(s.ExactCache[key], sub)

		// Evict cache entry if it exceeds the cache size limit
		if len(s.ExactCache) > s.cacheSize {
			for k := range s.ExactCache {
				delete(s.ExactCache, k)
				x_log.Info("Evicted exact Subscription from cache: subject =", k)
				break
			}
		}

		x_log.Info("Inserted exact Subscription: subject =", key)
	}
}

//-------------------------------------------------
// Remove - Deletes a Subscription
//-------------------------------------------------

// Remove deletes a Subscription from the tree.
func (s *Sublist) Remove(sub *typ.Subscription) error {
	key := string(sub.Subject)
	x_log.Info("Removing subscription: subject =", key)

	// Check if the subscription exists in the tree
	existingSub, exists := s.Tree.Find(sub.Subject)
	if !exists || existingSub == nil {
		x_log.Error("Subscription with subject", key, "not found in tree")
		return fmt.Errorf("subscription with subject %s not found in tree", key)
	}

	// Try to remove the subscription from the tree
	s.Tree.Delete(sub.Subject)

	// Remove from the appropriate cache (exact or wildcard)
	if isWildcard(key) {
		if _, exists := s.WildcardCache[key]; !exists {
			x_log.Error("Wildcard subscription with subject", key, "not found in cache")
			return fmt.Errorf("wildcard subscription with subject %s not found in cache", key)
		}
		delete(s.WildcardCache, key)
		x_log.Info("Removed wildcard Subscription: subject =", key)
	} else {
		if _, exists := s.ExactCache[key]; !exists {
			x_log.Error("Exact subscription with subject", key, "not found in cache")
			return fmt.Errorf("exact subscription with subject %s not found in cache", key)
		}
		delete(s.ExactCache, key)
		x_log.Info("Removed exact Subscription: subject =", key)
	}

	return nil
}

//-------------------------------------------------
// SublistResult - Holds the subscriptions that match a subject
//-------------------------------------------------

// SublistResult holds the subscriptions that match a subject.
type SublistResult struct {
	Psubs []*typ.Subscription // List of matching subscriptions
}

//-------------------------------------------------
// Match - Returns Subscriptions matching the given subject
//-------------------------------------------------

// Match returns Subscriptions matching the given subject.
func (s *Sublist) Match(subject []byte) *SublistResult {
	x_log.Info("Matching subscriptions for subject:", string(subject))
	res := &SublistResult{}

	// Debugging the current subscriptions
	x_log.Debug("Exact cache:", s.ExactCache)
	x_log.Debug("Wildcard cache:", s.WildcardCache)

	// Handle regular subjects
	if subs, ok := s.ExactCache[string(subject)]; ok {
		x_log.Info("Exact match found for subject:", string(subject))
		res.Psubs = append(res.Psubs, subs...)
	}

	// Use the tree to match subscriptions (e.g., with wildcards)
	s.Tree.Match(subject, func(_ []byte, sub **typ.Subscription) {
		if sub != nil && *sub != nil && !contains(res.Psubs, *sub) {
			res.Psubs = append(res.Psubs, *sub)
			x_log.Info("Matching subscription added: subject =", string(subject))
		}
	})

	x_log.Info("Tree matched Subscriptions: count =", len(res.Psubs), ", subject =", string(subject))
	return res
}

//-------------------------------------------------
// HasInterest - Checks if any subscription matches the subject
//-------------------------------------------------

// HasInterest checks if any subscription matches the subject.
func (s *Sublist) HasInterest(subject []byte) bool {
	x_log.Info("Checking interest for subject:", string(subject))
	found := false
	s.Tree.Match(subject, func(_ []byte, _ **typ.Subscription) {
		found = true
	})
	x_log.Info("HasInterest: subject =", string(subject), ", found =", found)
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
func contains(list []*typ.Subscription, target *typ.Subscription) bool {
	for _, s := range list {
		if s == target {
			x_log.Debug("Subscription already in list:", s)
			return true
		}
	}
	x_log.Debug("Subscription not in list:", target)
	return false
}

//-------------------------------------------------
// MatchSubjectWithWildcard - Matching subscriptions with wildcard filter
//-------------------------------------------------

// MatchSubjectWithWildcard applies wildcard matching to subscriptions.
func (s *Sublist) MatchSubjectWithWildcard(filter []byte) *SublistResult {
	res := &SublistResult{}

	// Use tree.Match to find subscriptions matching the wildcard filter
	s.Tree.Match(filter, func(subject []byte, sub **typ.Subscription) {
		if sub != nil && *sub != nil {
			res.Psubs = append(res.Psubs, *sub) // Add the matched subscription to the result
		}
	})

	return res
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
	x_log.Info("Creating subject transform: src =", src, ", dest =", dest)
	srcTokens := strings.Split(src, ".")
	destTokens := strings.Split(dest, ".")

	// Check if the number of wildcards matches between source and destination
	if countWildcards(srcTokens) != countWildcards(destTokens) {
		x_log.Error("Wildcard count mismatch in transform: src =", src, ", dest =", dest)
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
	x_log.Info("Transforming subject:", subject)
	inputTokens := strings.Split(subject, ".")
	mapping := make([]string, 0)
	i := 0

	// Apply the transformation rules to the source subject
	for _, token := range st.srcTokens {
		switch token {
		case "*":
			if i >= len(inputTokens) {
				x_log.Error("Subject too short for *: subject =", subject)
				return "", errors.New("subject too short for *")
			}
			mapping = append(mapping, inputTokens[i])
			i++
		case ">":
			if i >= len(inputTokens) {
				x_log.Error("No tokens available for >: subject =", subject)
				return "", errors.New("no tokens available for >")
			}
			mapping = append(mapping, strings.Join(inputTokens[i:], "."))
			i = len(inputTokens) // Move the index to the end
		default:
			if i >= len(inputTokens) || token != inputTokens[i] {
				x_log.Error("Subject does not match pattern: expected =", token, ", got =", inputTokens[i])
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
				x_log.Error("Not enough wildcards for destination: subject =", subject)
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
	x_log.Info("Transformed subject: input =", subject, ", output =", transformed)
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
	x_log.Info("Matching subscriptions for subject:", string(subject), "and queue:", string(queue))
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
	s.Tree.Match(subject, func(_ []byte, sub **typ.Subscription) {
		if sub != nil && *sub != nil && string((*sub).Queue) == string(queue) {
			res.Psubs = append(res.Psubs, *sub)
		}
	})

	x_log.Info("Matched", len(res.Psubs), "subscriptions for subject", string(subject), "and queue", string(queue))
	return res
}
