package session

type FeedbackEnqueuer interface {
	EnqueueFeedback(context, code, selected, top string) bool
	RecordSelectedUsage(text string)
}

type EngineFeedbackSink struct {
	Target FeedbackEnqueuer
}

func (s EngineFeedbackSink) AcceptFeedbackEvents(events []FeedbackEvent) {
	if s.Target == nil {
		return
	}
	for _, evt := range events {
		_ = s.Target.EnqueueFeedback(evt.EffectiveContext, evt.InputCode, evt.Selected.Text, evt.Top.Text)
		s.Target.RecordSelectedUsage(evt.Selected.Text)
	}
}
