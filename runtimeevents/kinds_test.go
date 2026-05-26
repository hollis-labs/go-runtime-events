package runtimeevents

import "testing"

func TestPolicyApprovalRequestedKind(t *testing.T) {
	if KindPolicyApprovalRequested != EventKind("policy.approval_requested") {
		t.Errorf("KindPolicyApprovalRequested = %q", KindPolicyApprovalRequested)
	}
}
