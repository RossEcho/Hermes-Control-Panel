package hermes_test

import (
	"context"
	"testing"

	"github.com/RossEcho/hermes-control-panel/internal/hermes"
)

func newMock() *hermes.MockAdapter {
	return hermes.NewMockAdapter()
}

func TestListSkillsNonNil(t *testing.T) {
	m := newMock()
	skills, err := m.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills: %v", err)
	}
	if len(skills) == 0 {
		t.Error("expected at least one skill")
	}
}

func TestGetSkill(t *testing.T) {
	m := newMock()
	skills, _ := m.ListSkills()
	if len(skills) == 0 {
		t.Skip("no skills to test")
	}
	skill, err := m.GetSkill(skills[0].ID)
	if err != nil {
		t.Fatalf("GetSkill: %v", err)
	}
	if skill == nil {
		t.Fatal("expected non-nil skill")
	}
	if skill.ID != skills[0].ID {
		t.Errorf("ID mismatch: %q != %q", skill.ID, skills[0].ID)
	}
}

func TestGetSkillNotFound(t *testing.T) {
	m := newMock()
	_, err := m.GetSkill("nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent skill, got nil")
	}
}

func TestListModels(t *testing.T) {
	m := newMock()
	models, err := m.ListModels()
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) < 2 {
		t.Error("expected at least 2 models")
	}
	// Exactly one should be active
	active := 0
	for _, mod := range models {
		if mod.IsActive {
			active++
		}
	}
	if active != 1 {
		t.Errorf("expected exactly 1 active model, got %d", active)
	}
}

func TestSwitchModel(t *testing.T) {
	m := newMock()
	models, _ := m.ListModels()
	// Find a non-active model
	var target hermes.Model
	for _, mod := range models {
		if !mod.IsActive {
			target = mod
			break
		}
	}
	if target.ID == "" {
		t.Skip("no inactive model to switch to")
	}
	if err := m.SwitchModel(target.ID); err != nil {
		t.Fatalf("SwitchModel: %v", err)
	}
	models, _ = m.ListModels()
	for _, mod := range models {
		if mod.ID == target.ID && !mod.IsActive {
			t.Errorf("model %q should now be active", target.ID)
		}
	}
}

func TestSwitchModelNotFound(t *testing.T) {
	m := newMock()
	if err := m.SwitchModel("no-such-model"); err == nil {
		t.Error("expected error for unknown model")
	}
}

func TestListSessions(t *testing.T) {
	m := newMock()
	sessions, err := m.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) == 0 {
		t.Error("expected pre-seeded sessions")
	}
}

func TestNewSession(t *testing.T) {
	m := newMock()
	sess, err := m.NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if sess == nil {
		t.Fatal("expected non-nil session")
	}
	if sess.ID == "" {
		t.Error("expected non-empty session ID")
	}
}

func TestGetSession(t *testing.T) {
	m := newMock()
	sessions, _ := m.ListSessions()
	if len(sessions) == 0 {
		t.Skip("no sessions")
	}
	s, err := m.GetSession(sessions[0].ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil session")
	}
}

func TestGetStatus(t *testing.T) {
	m := newMock()
	status, err := m.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.HermesProcess == "" {
		t.Error("expected non-empty HermesProcess")
	}
}

func TestListJobs(t *testing.T) {
	m := newMock()
	jobs, err := m.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) == 0 {
		t.Error("expected at least one job")
	}
}

func TestListProcesses(t *testing.T) {
	m := newMock()
	procs, err := m.ListProcesses()
	if err != nil {
		t.Fatalf("ListProcesses: %v", err)
	}
	if len(procs) == 0 {
		t.Error("expected at least one process")
	}
}

func TestGetLogs(t *testing.T) {
	m := newMock()
	logs, err := m.GetLogs(10)
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if len(logs) == 0 {
		t.Error("expected log lines")
	}
}

func TestGetLogsLimitsClamped(t *testing.T) {
	m := newMock()
	logs, err := m.GetLogs(1000) // more than available
	if err != nil {
		t.Fatalf("GetLogs: %v", err)
	}
	if len(logs) == 0 {
		t.Error("expected at least one log line")
	}
}

func TestRunAction(t *testing.T) {
	m := newMock()
	out, err := m.RunAction("test-action", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("RunAction: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output")
	}
}

func TestStreamChat(t *testing.T) {
	m := newMock()

	// Need a real session to stream into
	sess, _ := m.NewSession()

	ctx := context.Background()
	ch, err := m.StreamChat(ctx, sess.ID, "hello test")
	if err != nil {
		t.Fatalf("StreamChat: %v", err)
	}

	var events []hermes.Event
	for e := range ch {
		events = append(events, e)
	}

	if len(events) == 0 {
		t.Error("expected at least one event from StreamChat")
	}

	// Check we got a message_start and message_end
	hasStart, hasEnd := false, false
	for _, e := range events {
		if e.Type == "message_start" {
			hasStart = true
		}
		if e.Type == "message_end" {
			hasEnd = true
		}
	}
	if !hasStart {
		t.Error("expected message_start event")
	}
	if !hasEnd {
		t.Error("expected message_end event")
	}
}

func TestStreamChatBadSession(t *testing.T) {
	m := newMock()
	_, err := m.StreamChat(context.Background(), "no-such-session", "hello")
	if err == nil {
		t.Error("expected error for bad session ID")
	}
}

func TestStopChat(t *testing.T) {
	m := newMock()
	// StopChat on a session with no active stream should not error
	if err := m.StopChat("sess-001"); err != nil {
		t.Errorf("StopChat: %v", err)
	}
}

func TestGetSessionMessages(t *testing.T) {
	m := newMock()
	sessions, _ := m.ListSessions()
	if len(sessions) == 0 {
		t.Skip("no sessions")
	}
	msgs, err := m.GetSessionMessages(sessions[0].ID)
	if err != nil {
		t.Fatalf("GetSessionMessages: %v", err)
	}
	// Pre-seeded sessions start empty (messages accumulate only after chat)
	_ = msgs
}

func TestGetSessionMessagesBadSession(t *testing.T) {
	m := newMock()
	_, err := m.GetSessionMessages("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}
