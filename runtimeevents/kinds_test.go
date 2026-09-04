package runtimeevents

import "testing"

func TestPolicyCompatibilityKinds(t *testing.T) {
	tests := []struct {
		name string
		got  EventKind
		want string
	}{
		{"KindPolicyNudge", KindPolicyNudge, "policy.nudge"},
		{"KindPolicyRewrite", KindPolicyRewrite, "policy.rewrite"},
		{"KindPolicyBlock", KindPolicyBlock, "policy.block"},
		{"KindPolicyApprovalRequested", KindPolicyApprovalRequested, "policy.approval_requested"},
	}
	for _, tt := range tests {
		if string(tt.got) != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}
