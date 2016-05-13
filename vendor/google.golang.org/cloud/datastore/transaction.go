// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	pb "google.golang.org/cloud/internal/datastore"
)

// ErrConcurrentTransaction is returned when a transaction is rolled back due
// to a conflict with a concurrent transaction.
var ErrConcurrentTransaction = errors.New("datastore: concurrent transaction")

func runOnce(ctx context.Context, f func(context.Context) error) error {
	req := &pb.BeginTransactionRequest{}
	resp := &pb.BeginTransactionResponse{}
	if err := call(ctx, "beginTransaction", req, resp); err != nil {
		return err
	}
	subCtx := context.WithValue(ctx, ContextKey("transaction"), resp.Transaction)
	finished := false
	// Call f, rolling back the transaction if f returns a non-nil error, or panics.
	// The panic is not recovered.
	defer func() {
		if finished {
			return
		}
		finished = true
		// Ignore the error return value, since we are already returning a non-nil
		// error (or we're panicking).
		call(subCtx, "rollback", &pb.RollbackRequest{Transaction: resp.Transaction}, &pb.RollbackResponse{})
	}()
	if err := f(subCtx); err != nil {
		return err
	}
	finished = true
	err := call(subCtx, "commit", &pb.CommitRequest{Transaction: resp.Transaction}, &pb.CommitResponse{})
	if e, ok := err.(*errHTTP); ok && e.StatusCode == http.StatusConflict {
		// TODO(jbd): Make sure that we explicitly handle the case where response
		// has an HTTP 409 and the error message indicates that it's an concurrent
		// transaction error.
		return ErrConcurrentTransaction
	}
	return err
}

// RunInTransaction runs f in a transaction. It calls f with a transaction
// context tc that f should use for all App Engine operations.
//
// If f returns nil, RunInTransaction attempts to commit the transaction,
// returning nil if it succeeds. If the commit fails due to a conflicting
// transaction, RunInTransaction retries f, each time with a new transaction
// context. It gives up and returns ErrConcurrentTransaction after three
// failed attempts.
//
// If f returns non-nil, then any datastore changes will not be applied and
// RunInTransaction returns that same error. The function f is not retried.
//
// Note that when f returns, the transaction is not yet committed. Calling code
// must be careful not to assume that any of f's changes have been committed
// until RunInTransaction returns nil.
//
// Nested transactions are not supported; c may not be a transaction context.
func RunInTransaction(ctx context.Context, f func(context.Context) error) error {
	if transaction(ctx) != nil {
		return errors.New("datastore: nested transactions are not supported")
	}
	for i := 0; i < 3; i++ {
		if err := runOnce(ctx, f); err != ErrConcurrentTransaction {
			return err
		}
	}
	return ErrConcurrentTransaction
}
